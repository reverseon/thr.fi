package url

import (
	"context"
	"fmt"
	"net/http"
	"urlshortener/commons"

	"github.com/gin-gonic/gin"
	"github.com/supertokens/supertokens-golang/recipe/session"
)

type URLController struct {
	url_service *URLService
}

// CREATE

type CreateShortenedURLJSONInput struct {
	Original_url string `json:"original_url" binding:"required,url"`
	Backhalf     string `json:"backhalf" binding:"omitempty,alphanum,min=1,max=20"`
	Password     string `json:"password" binding:"omitempty,min=8,max=32"`
}

func (ctl *URLController) createShortenedURL(c *gin.Context) {
	// variable declaration
	var http_input CreateShortenedURLJSONInput
	var err error
	var service_error *commons.ServiceError
	var returned_from_service CreateShortenedURLOutput

	// input validation
	if err = c.ShouldBindJSON(&http_input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// service input construction
	var service_input = struct {
		request_context context.Context
		original_url    string
		backhalf        *string
		password        *string
		creator_id      *string
	}{
		request_context: c.Request.Context(),
		original_url:    http_input.Original_url,
		backhalf:        nil,
		password:        nil,
		creator_id:      nil,
	}
	if http_input.Backhalf != "" {
		service_input.backhalf = &http_input.Backhalf
	}
	if http_input.Password != "" {
		service_input.password = &http_input.Password
	}
	session_container, err := session.GetSession(c.Request, c.Writer, nil)
	if err != nil {
		// session does not exist
		service_input.creator_id = nil
	} else {
		// session exists
		var user_id string = session_container.GetUserID()
		service_input.creator_id = &user_id
	}

	// service call
	returned_from_service, service_error = ctl.url_service.createShortenedURL(
		service_input.request_context,
		service_input.original_url,
		service_input.backhalf,
		service_input.password,
		service_input.creator_id,
	)
	if service_error != nil {
		switch service_error.Code {
		case "BACKHALF_EXISTS":
			c.JSON(http.StatusConflict, gin.H{"error": "Backhalf already exists"})
		case "REDIS_ERROR":
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Something went wrong on our end"})
		case "BACKHALF_GENERATION_ERROR":
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Something went wrong on our end"})
		default:
			fmt.Println("Unhandled error code: " + service_error.Code)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Something went wrong on our end"})
		}
		return
	} else {
		c.JSON(http.StatusCreated, gin.H{
			"original_url":       returned_from_service.Original_url,
			"backhalf":           returned_from_service.Backhalf,
			"password_protected": returned_from_service.Password_protected,
		})
		return
	}
}

// READ

func (ctl *URLController) getURLInfoByBackhalf(c *gin.Context) {
	// variable declaration
	var backhalf string
	var service_error *commons.ServiceError
	var returned_from_service GetShortenedURLOutputByBackhalf

	// input validation
	backhalf = c.Param("backhalf")

	// service input construction
	var service_input = struct {
		request_context context.Context
		requester_id    *string
		backhalf        string
	}{
		request_context: c.Request.Context(),
		requester_id:    nil,
		backhalf:        backhalf,
	}

	session_container, err := session.GetSession(c.Request, c.Writer, nil)
	if err != nil {
		// session does not exist
		service_input.requester_id = nil
	} else {
		// session exists
		var user_id string = session_container.GetUserID()
		service_input.requester_id = &user_id
	}

	// service call
	returned_from_service, service_error = ctl.url_service.getShortenedURLByBackhalf(
		service_input.request_context,
		service_input.backhalf,
		service_input.requester_id,
		nil,
	)
	if service_error != nil {
		switch service_error.Code {
		case "BACKHALF_NOT_EXISTS":
			c.JSON(http.StatusNotFound, gin.H{"error": "URL not found"})
		case "PASSWORD_REQUIRED":
			c.JSON(http.StatusForbidden, gin.H{"error": "Password required"})
		case "WRONG_PASSWORD":
			c.JSON(http.StatusForbidden, gin.H{"error": "Wrong password"})
		case "REDIS_ERROR":
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Something went wrong on our end"})
		default:
			fmt.Println("Unhandled error code: " + service_error.Code)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Something went wrong on our end"})
		}
		return
	} else {
		c.JSON(http.StatusOK, gin.H{
			"original_url":       returned_from_service.Original_url,
			"backhalf":           returned_from_service.Backhalf,
			"password_protected": returned_from_service.Password_protected,
		})
		return
	}
}

func (ctl *URLController) getUserURLs(c *gin.Context) {
	// variable declaration
	var service_error *commons.ServiceError
	var returned_from_service GetShortenedURLOutputByUser

	// get session (this use verifySession middleware)
	var user_id string = session.GetSessionFromRequestContext(c.Request.Context()).GetUserID()

	// get page
	var page string = c.DefaultQuery("page", "1")
	var per_page string = c.DefaultQuery("per_page", "10")

	// convert page and per_page to int
	var page_int, per_page_int int
	_, err := fmt.Sscanf(page, "%d", &page_int)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid page"})
		return
	}
	_, err = fmt.Sscanf(per_page, "%d", &per_page_int)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid per_page"})
		return
	}

	// service input construction
	var service_input = struct {
		request_context context.Context
		requester_id    string
		user_id         string
	}{
		request_context: c.Request.Context(),
		requester_id:    user_id,
		user_id:         user_id,
	}

	// service call
	returned_from_service, service_error = ctl.url_service.getShortenedURLByUser(
		service_input.request_context,
		service_input.user_id,
		service_input.requester_id,
		page_int,
		int64(per_page_int),
	)
	if service_error != nil {
		switch service_error.Code {
		case "REDIS_ERROR":
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Something went wrong on our end"})
		case "INVALID_PAGE":
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid page, must be greater than 0"})
		case "INVALID_PER_PAGE":
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid per_page, must be greater than 0"})
		default:
			fmt.Println("Unhandled error code: " + service_error.Code)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Something went wrong on our end"})
		}
		return
	} else {
		c.JSON(http.StatusOK, gin.H{
			"total": returned_from_service.total,
			"urls":  returned_from_service.data,
		})
		return
	}
}

type UnlockURLJSONInput struct {
	Password string `json:"password" binding:"required,min=8,max=32"`
}

func (ctl *URLController) unlockURL(c *gin.Context) {
	// variable declaration
	var backhalf string
	var http_input UnlockURLJSONInput
	var err error
	var service_error *commons.ServiceError
	var returned_from_service GetShortenedURLOutputByBackhalf

	// input validation
	backhalf = c.Param("backhalf")
	if err = c.ShouldBindJSON(&http_input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// service input construction
	var service_input = struct {
		request_context context.Context
		backhalf        string
		password        string
	}{
		request_context: c.Request.Context(),
		backhalf:        backhalf,
		password:        http_input.Password,
	}

	// service call
	returned_from_service, service_error = ctl.url_service.getShortenedURLByBackhalf(
		service_input.request_context,
		service_input.backhalf,
		nil,
		&service_input.password,
	)
	if service_error != nil {
		switch service_error.Code {
		case "BACKHALF_NOT_EXISTS":
			c.JSON(http.StatusNotFound, gin.H{"error": "URL not found"})
		case "PASSWORD_REQUIRED":
			c.JSON(http.StatusForbidden, gin.H{"error": "Password required"})
		case "WRONG_PASSWORD":
			c.JSON(http.StatusForbidden, gin.H{"error": "Wrong password"})
		case "REDIS_ERROR":
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Something went wrong on our end"})
		default:
			fmt.Println("Unhandled error code: " + service_error.Code)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Something went wrong on our end"})
		}
		return
	} else {
		c.JSON(http.StatusOK, gin.H{
			"original_url":       returned_from_service.Original_url,
			"backhalf":           returned_from_service.Backhalf,
			"password_protected": returned_from_service.Password_protected,
		})
		return
	}
}

// UPDATE

type UpdateURLJSONInput struct {
	Backhalf     string `json:"backhalf" binding:"omitempty,alphanum,min=1,max=20"`
	Password     string `json:"password" binding:"omitempty,min=8,max=32"`
	Original_url string `json:"original_url" binding:"omitempty,url"`
}

func (ctl *URLController) updateURL(c *gin.Context) {
	// variable declaration
	var backhalf string
	var http_input UpdateURLJSONInput
	var err error
	var service_error *commons.ServiceError
	var returned_from_service UpdateShortenedURLOutput

	// input validation
	backhalf = c.Param("backhalf")
	if err = c.ShouldBindJSON(&http_input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// service input construction
	var service_input = struct {
		request_context   context.Context
		original_backhalf string
		updater_id        string
		changed_backhalf  *string
		original_url      *string
		password          *string
	}{
		request_context:   c.Request.Context(),
		original_backhalf: backhalf,
		updater_id:        session.GetSessionFromRequestContext(c.Request.Context()).GetUserID(),
		changed_backhalf:  nil,
		original_url:      nil,
		password:          nil,
	}

	if http_input.Backhalf != "" {
		service_input.changed_backhalf = &http_input.Backhalf
	}
	if http_input.Password != "" {
		service_input.password = &http_input.Password
	}
	if http_input.Original_url != "" {
		service_input.original_url = &http_input.Original_url
	}

	// service call
	returned_from_service, service_error = ctl.url_service.updateShortenedURL(
		service_input.request_context,
		service_input.original_backhalf,
		service_input.updater_id,
		service_input.changed_backhalf,
		service_input.original_url,
		service_input.password,
	)
	if service_error != nil {
		switch service_error.Code {
		case "BACKHALF_NOT_EXISTS":
			c.JSON(http.StatusNotFound, gin.H{"error": "URL not found"})
		case "SAME_USER_REQUIRED":
			c.JSON(http.StatusForbidden, gin.H{"error": "You are not the creator of this URL"})
		case "BACKHALF_EXISTS":
			c.JSON(http.StatusConflict, gin.H{"error": "Backhalf already exists"})
		case "REDIS_ERROR":
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Something went wrong on our end"})
		default:
			fmt.Println("Unhandled error code: " + service_error.Code)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Something went wrong on our end"})
		}
		return
	} else {
		c.JSON(http.StatusOK, gin.H{
			"original_url":       returned_from_service.Original_url,
			"backhalf":           returned_from_service.Backhalf,
			"password_protected": returned_from_service.Password_protected,
		})
		return
	}

}

func (ctl *URLController) disablePasswordProtection(c *gin.Context) {
	// variable declaration
	var backhalf string
	var service_error *commons.ServiceError
	var returned_from_service DisablePasswordProtectionOutput

	// input validation
	backhalf = c.Param("backhalf")

	// service input construction
	var service_input = struct {
		request_context context.Context
		backhalf        string
		updater_id      string
	}{
		request_context: c.Request.Context(),
		backhalf:        backhalf,
		updater_id:      session.GetSessionFromRequestContext(c.Request.Context()).GetUserID(),
	}

	// service call
	returned_from_service, service_error = ctl.url_service.disablePasswordProtection(
		service_input.request_context,
		service_input.backhalf,
		service_input.updater_id,
	)
	if service_error != nil {
		switch service_error.Code {
		case "BACKHALF_NOT_EXISTS":
			c.JSON(http.StatusNotFound, gin.H{"error": "URL not found"})
		case "SAME_USER_REQUIRED":
			c.JSON(http.StatusForbidden, gin.H{"error": "You are not the creator of this URL"})
		case "REDIS_ERROR":
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Something went wrong on our end"})
		default:
			fmt.Println("Unhandled error code: " + service_error.Code)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Something went wrong on our end"})
		}
		return
	} else {
		c.JSON(http.StatusOK, gin.H{
			"original_url":       returned_from_service.Original_url,
			"backhalf":           returned_from_service.Backhalf,
			"password_protected": returned_from_service.Password_protected,
		})
		return
	}
}

func (ctl *URLController) deleteURL(c *gin.Context) {
	// variable declaration
	var backhalf string
	var service_error *commons.ServiceError
	var returned_from_service DeleteShortenedURLOutput

	// input validation
	backhalf = c.Param("backhalf")

	// service input construction
	var service_input = struct {
		request_context context.Context
		backhalf        string
		deleter_id      string
	}{
		request_context: c.Request.Context(),
		backhalf:        backhalf,
		deleter_id:      session.GetSessionFromRequestContext(c.Request.Context()).GetUserID(),
	}

	// service call
	returned_from_service, service_error = ctl.url_service.deleteShortenedURL(
		service_input.request_context,
		service_input.backhalf,
		service_input.deleter_id,
	)
	if service_error != nil {
		switch service_error.Code {
		case "BACKHALF_NOT_EXISTS":
			c.JSON(http.StatusNotFound, gin.H{"error": "URL not found"})
		case "SAME_USER_REQUIRED":
			c.JSON(http.StatusForbidden, gin.H{"error": "You are not the creator of this URL"})
		case "REDIS_ERROR":
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Something went wrong on our end"})
		default:
			fmt.Println("Unhandled error code: " + service_error.Code)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Something went wrong on our end"})
		}
		return
	} else {
		c.JSON(http.StatusOK, gin.H{
			"original_url":       returned_from_service.Original_url,
			"backhalf":           returned_from_service.Backhalf,
			"password_protected": returned_from_service.Password_protected,
		})
		return
	}
}

func (ctl *URLController) ApplyRouterGroup(group *gin.RouterGroup) {
	group.POST("/create", ctl.createShortenedURL)
	group.GET("/:backhalf", ctl.getURLInfoByBackhalf)
	group.GET("/user", commons.VerifySession(nil), ctl.getUserURLs)
	group.POST("/:backhalf/unlock", ctl.unlockURL)
	group.PUT("/:backhalf", commons.VerifySession(nil), ctl.updateURL)
	group.PUT("/:backhalf/disable_password", commons.VerifySession(nil), ctl.disablePasswordProtection)
	group.DELETE("/:backhalf", commons.VerifySession(nil), ctl.deleteURL)
}
