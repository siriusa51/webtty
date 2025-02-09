package apis

import (
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

type JSONResponse map[string]any

func writeJSONResponse(ctx *gin.Context, code int, obj any) {
	ctx.JSON(code, obj)
}

func writeWebSocketJSONResponse(conn *websocket.Conn, obj any) {
	conn.WriteJSON(obj)
}
