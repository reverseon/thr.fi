package user

import (
	"net/http"
	"urlshortener/commons"

	"github.com/gin-gonic/gin"
	"github.com/supertokens/supertokens-golang/recipe/emailpassword/epmodels"
	"github.com/supertokens/supertokens-golang/recipe/session"
)

type UserController struct {
	auth_service *UserService
}

func (ctl *UserController) getUserInfo(c *gin.Context) {
	var user_id string = session.GetSessionFromRequestContext(c.Request.Context()).GetUserID()
	var returned *epmodels.User
	var err *commons.ServiceError
	returned, err = ctl.auth_service.getUserInfo(user_id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Code})
	} else {
		c.JSON(http.StatusOK, gin.H{
			"id":    returned.ID,
			"email": returned.Email,
		})
	}

}

// exported

func (ctl *UserController) ApplyRouterGroup(group *gin.RouterGroup) {
	group.GET("/info", commons.VerifySession(nil), ctl.getUserInfo)
}
