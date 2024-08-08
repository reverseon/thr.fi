package main

import (
	"net/http"
	"os"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"

	"urlshortener/url"
	"urlshortener/user"

	"github.com/supertokens/supertokens-golang/recipe/emailpassword"
	"github.com/supertokens/supertokens-golang/recipe/emailpassword/epmodels"
	"github.com/supertokens/supertokens-golang/recipe/session"
	"github.com/supertokens/supertokens-golang/recipe/session/sessmodels"
	"github.com/supertokens/supertokens-golang/supertokens"
)

func applyRouter(r *gin.Engine) {
	var user_controller user.UserController = user.UserController{}
	var url_controller url.URLController = url.URLController{}
	// User Routes
	user_controller.ApplyRouterGroup(r.Group("/user"))
	// URL Routes
	url_controller.ApplyRouterGroup(r.Group("/url"))
	// Default Health Check
	r.GET("/", func(c *gin.Context) {
		c.String(http.StatusOK, "OK!")
	})
}

func main() {
	var err error
	err = godotenv.Load()
	if err != nil {
		panic("Error loading .env file")
	}
	var api_base_path string = "/auth"
	var website_base_path string = "/auth"
	var api_domain string = os.Getenv("API_DOMAIN")
	var allowed_frontend_domain = os.Getenv("ALLOWED_FRONTEND_DOMAIN")
	var supertokens_core_api_url string = os.Getenv("SUPERTOKENS_CORE_API_URL")
	var supertokens_core_api_key string = os.Getenv("SUPERTOKENS_CORE_API_KEY")
	var app_name = "URL Shortener"
	err = supertokens.Init(supertokens.TypeInput{
		Supertokens: &supertokens.ConnectionInfo{
			ConnectionURI: supertokens_core_api_url,
			APIKey:        supertokens_core_api_key,
		},
		AppInfo: supertokens.AppInfo{
			AppName:         app_name,
			APIDomain:       api_domain,
			WebsiteDomain:   allowed_frontend_domain,
			APIBasePath:     &api_base_path,
			WebsiteBasePath: &website_base_path,
		},
		RecipeList: []supertokens.Recipe{
			emailpassword.Init(&epmodels.TypeInput{
				Override: &epmodels.OverrideStruct{
					APIs: func(originalImplementation epmodels.APIInterface) epmodels.APIInterface {
						originalImplementation.SignUpPOST = nil
						originalImplementation.SignInPOST = nil
						originalImplementation.EmailExistsGET = nil
						originalImplementation.GeneratePasswordResetTokenPOST = nil
						originalImplementation.PasswordResetPOST = nil
						return originalImplementation
					},
				},
			}),
			session.Init(&sessmodels.TypeInput{
				GetTokenTransferMethod: func(req *http.Request, forCreateNewSession bool, userContext supertokens.UserContext) sessmodels.TokenTransferMethod {
					return sessmodels.HeaderTransferMethod
				},
			}),
		},
	})
	if err != nil {
		panic(err.Error())
	}

	var r *gin.Engine = gin.Default()
	r.Use(cors.New(cors.Config{
		AllowOrigins: []string{allowed_frontend_domain},
		AllowMethods: []string{"GET", "POST", "DELETE", "PUT", "OPTIONS"},
		AllowHeaders: append([]string{"content-type"},
			supertokens.GetAllCORSHeaders()...),
		AllowCredentials: true,
	}))

	r.Use(func(c *gin.Context) {
		supertokens.Middleware(http.HandlerFunc(
			func(rw http.ResponseWriter, r *http.Request) {
				c.Next()
			})).ServeHTTP(c.Writer, c.Request)
		// we call Abort so that the next handler in the chain is not called, unless we call Next explicitly
		c.Abort()
	})
	applyRouter(r)
	r.Run("0.0.0.0:6173")

}
