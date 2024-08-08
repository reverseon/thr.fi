package user

import (
	"urlshortener/commons"

	"github.com/supertokens/supertokens-golang/recipe/emailpassword"
	"github.com/supertokens/supertokens-golang/recipe/emailpassword/epmodels"
)

type UserService struct{}

func (svc *UserService) getUserInfo(user_id string) (*epmodels.User, *commons.ServiceError) {
	var user_info *epmodels.User
	var err error
	user_info, err = emailpassword.GetUserByID(user_id)
	if err != nil {
		return nil, &commons.ServiceError{Code: "INTERNAL_SERVER_ERROR"}
	} else {
		return user_info, nil
	}
}
