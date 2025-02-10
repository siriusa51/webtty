package apis

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/siriusa51/webtty/session"
	"github.com/siriusa51/webtty/tty"
	"golang.org/x/sync/errgroup"
	"io"
	"log/slog"
	"net/http"
)

var SessionStopped = errors.New("session stopped")

type ControllerConfig struct {
	Workdir  string
	Command  string
	ExtraEnv []string
}

type Controller struct {
	log      *slog.Logger
	config   ControllerConfig
	mgr      *session.SessionManager
	upgrader *websocket.Upgrader
}

func NewController(config ControllerConfig, log *slog.Logger, mgr *session.SessionManager) *Controller {
	return &Controller{
		config: config,
		log:    log.With("module", "apis/controller"),
		mgr:    mgr,
		upgrader: &websocket.Upgrader{
			ReadBufferSize:  4096,
			WriteBufferSize: 4096,
			CheckOrigin:     func(r *http.Request) bool { return true },
		},
	}
}

// Websocket handles the websocket connection for tty.
func (c *Controller) Websocket(ctx *gin.Context) {
	sid := ctx.Query("sid")
	if sid == "" {
		writeJSONResponse(ctx, http.StatusBadRequest, JSONResponse{"error": "sid is required"})
		return
	}

	conn, err := c.upgrader.Upgrade(ctx.Writer, ctx.Request, nil)
	if err != nil {
		c.log.Error("failed to upgrade connection", "error", err)
		writeJSONResponse(ctx, http.StatusBadRequest, JSONResponse{"error": "failed to upgrade connection"})
		return
	}

	defer conn.Close()

	session, err := c.mgr.GetSession(sid, func() (session.SessionIO, error) {
		return tty.New(c.config.Command,
			tty.WithWorkdir(c.config.Workdir),
			tty.WithContext(ctx),
			tty.WithExtraEnv(c.config.ExtraEnv...),
		)
	})

	if err != nil {
		c.log.Error("failed to get session", "error", err)
		writeWebSocketJSONResponse(conn, JSONResponse{"error": err.Error()})
		return
	}

	log := c.log.With("sid", sid)
	log.Info("session created")
	defer func() { log.Info("websocket closed") }()

	if err := session.Occupy(); err != nil {
		log.Error("failed to occupy session", "error", err)
		writeWebSocketJSONResponse(conn, JSONResponse{"error": err.Error()})
		return
	}

	defer session.Release()

	eg, egctx := errgroup.WithContext(ctx)
	eg.SetLimit(2)
	eg.Go(ttyClientHandler(egctx, log, conn, session))
	eg.Go(ttyServerHandler(egctx, log, conn, session))

	if err := eg.Wait(); err != nil {
		err = errors.Unwrap(err)
		switch err.(type) {
		case *websocket.CloseError:
			break
		default:
			if errors.Is(err, SessionStopped) {
				break
			}

			if errors.Is(err, io.EOF) {
				break
			}
			log.Error("failed to handle websocket", "error", err)
		}
		writeWebSocketJSONResponse(conn, JSONResponse{"error": err.Error()})
		return
	}
}

// RemoveSession removes the session by sid.
func (c *Controller) RemoveSession(ctx *gin.Context) {
	sid := ctx.Query("sid")
	c.mgr.RemoveSession(sid)
	writeJSONResponse(ctx, http.StatusOK, JSONResponse{"sid": sid})
}

// ttyClientHandler handles the client side of the tty.
// It reads the client and writes to the session.
func ttyClientHandler(ctx context.Context, log *slog.Logger, conn *websocket.Conn, sess *session.Session) func() error {
	return func() error {
		log.Info("tty client handler started")
		defer func() { log.Info("tty client handler stopped") }()

		for {
			select {
			case <-ctx.Done():
				return SessionStopped
			default:
				mt, message, err := conn.ReadMessage()
				if err != nil {
					return fmt.Errorf("failed to read message from client: %w", err)
				}

				if mt == websocket.CloseMessage {
					return SessionStopped
				}

				if mt != websocket.TextMessage {
					return fmt.Errorf("invalid message type: %d", mt)
				}

				if len(message) == 0 {
					return fmt.Errorf("empty message")
				}

				switch message[0] {
				case Input:
					if len(message) == 1 {
						continue
					}

					if _, err := io.Copy(sess, bytes.NewBuffer(message[1:])); err != nil {
						log.Warn("failed to write message to session", "error", err)
					}

				case ResizeTerminal:
					var resizeMessage ResizeMessage
					if err := json.Unmarshal(message[1:], &resizeMessage); err != nil {
						return fmt.Errorf("failed to unmarshal resize message: %w", err)
					}

					if err := sess.ResizeWindow(resizeMessage.Width, resizeMessage.Height); err != nil {
						return fmt.Errorf("failed to resize terminal: %w", err)
					}
				case Ping:
					pong := string(Pong) + "pong"
					if err := conn.WriteMessage(websocket.TextMessage, []byte(pong)); err != nil {
						return fmt.Errorf("failed to write pong message to client: %w", err)
					}
				default:
					return fmt.Errorf("invalid message type: %d", message[0])
				}
			}
		}
	}
}

// ttyServerHandler handles the server side of the tty.
// It reads the session and writes to the client.
func ttyServerHandler(ctx context.Context, log *slog.Logger, conn *websocket.Conn, sess *session.Session) func() error {
	return func() error {
		log.Info("tty server handler started")

		defer func() {
			log.Info("tty server handler stopped")

			data := string(Closed) + "session closed"
			if err := conn.WriteMessage(websocket.TextMessage, []byte(data)); err != nil {
				log.Warn("failed to write closed message to client", "error", err)
			}
		}()

		buff := make([]byte, 4096)

		for {
			select {
			case <-ctx.Done():
				return SessionStopped
			default:
				n, err := sess.Read(buff)
				if err != nil {
					return fmt.Errorf("failed to read message from session: %w", err)
				}

				data := string(Output) + base64.StdEncoding.EncodeToString(buff[:n])
				if err := conn.WriteMessage(websocket.TextMessage, []byte(data)); err != nil {
					return fmt.Errorf("failed to write message to client: %w", err)
				}
			}
		}

		return SessionStopped
	}
}
