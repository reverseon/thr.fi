package url

//lint:file-ignore U1000 Buggy linter in vscode

import (
	"context"
	"crypto/rand"
	"fmt"
	"math/big"
	"strings"
	"urlshortener/commons"
)

const (
	charset         = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	backhalf_rd_len = 8
)

func generateBackhalf() (string, error) {
	backhalfBuilder := strings.Builder{}
	backhalfBuilder.Grow(backhalf_rd_len) // Pre-allocate the required capacity

	for i := 0; i < backhalf_rd_len; i++ {
		randomIndex, err := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		if err != nil {
			return "", err
		}
		backhalfBuilder.WriteByte(charset[randomIndex.Int64()])
	}

	return backhalfBuilder.String(), nil
}

type URLService struct{}

// CREATE

type CreateShortenedURLOutput = commons.TypeReturnedURLInfo

func (svc *URLService) createShortenedURL(
	input_request_context context.Context,
	input_original_url string,
	input_backhalf *string, // nil if want to be randomly generated
	input_password *string, // nil if not password protected
	input_creator_id *string, // nil if guest-created
) (CreateShortenedURLOutput, *commons.ServiceError) {
	// variable declaration
	var redis_client = commons.GetRedisClient()
	var exists int64
	var err error
	var backhalf_to_write string

	// check if backhalf exists
	{
		if input_backhalf != nil {
			exists, err = redis_client.Exists(input_request_context,
				commons.BH_PREFIX+(*input_backhalf),
			).Result()
			if err != nil {
				fmt.Println(err)
				return CreateShortenedURLOutput{}, &commons.ServiceError{Code: "REDIS_ERROR"}
			}
			if exists == 1 {
				return CreateShortenedURLOutput{}, &commons.ServiceError{Code: "BACKHALF_EXISTS"}
			}
		}
	}

	// hash password
	{
		if input_password != nil {
			*input_password = commons.HashString(*input_password)
		}
	}

	// write to backhalf data
	{
		var data_to_write commons.TypeStoredURLInfo = commons.TypeStoredURLInfo{
			Original_url: input_original_url,
			Password:     input_password,
			Creator_id:   input_creator_id,
		}
		// find backhalf until it is unique
		if input_backhalf == nil {
			for {
				backhalf_to_write, err = generateBackhalf()
				if err != nil {
					fmt.Println(err)
					return CreateShortenedURLOutput{}, &commons.ServiceError{Code: "BACKHALF_GENERATION_ERROR"}
				}
				exists, err = redis_client.Exists(
					input_request_context,
					commons.BH_PREFIX+backhalf_to_write,
				).Result()
				if err != nil {
					fmt.Println(err)
					return CreateShortenedURLOutput{}, &commons.ServiceError{Code: "REDIS_ERROR"}
				}
				if exists == 0 {
					break
				}
			}
		} else {
			backhalf_to_write = *input_backhalf
		}
		err = redis_client.Set(input_request_context,
			commons.BH_PREFIX+(backhalf_to_write), data_to_write,
			0).Err()
		if err != nil {
			fmt.Println(err)
			return CreateShortenedURLOutput{}, &commons.ServiceError{Code: "REDIS_ERROR"}
		}
	}

	// write to user data
	{
		if input_creator_id != nil {
			// k,v = userprefix+creator_id, backhalf[]
			err = redis_client.SAdd(input_request_context,
				commons.USER_PREFIX+(*input_creator_id), backhalf_to_write,
			).Err()
			if err != nil {
				fmt.Println(err)
				return CreateShortenedURLOutput{}, &commons.ServiceError{Code: "REDIS_ERROR"}
			}
		}
	}
	return CreateShortenedURLOutput(commons.TypeReturnedURLInfo{
		Original_url:       input_original_url,
		Backhalf:           backhalf_to_write,
		Password_protected: input_password != nil,
	}), nil
}

// READ

type GetShortenedURLOutputByBackhalf = commons.TypeReturnedURLInfo

func (svc *URLService) getShortenedURLByBackhalf(
	input_request_context context.Context,
	input_backhalf string,
	input_requester_id *string,
	input_password *string,
) (GetShortenedURLOutputByBackhalf, *commons.ServiceError) {
	// variable declaration
	var redis_client = commons.GetRedisClient()
	var exists int64
	var err error
	var existing_data commons.TypeStoredURLInfo

	// check if backhalf exists
	{
		exists, err = redis_client.Exists(input_request_context, commons.BH_PREFIX+input_backhalf).Result()
		if err != nil {
			fmt.Println(err)
			return GetShortenedURLOutputByBackhalf{}, &commons.ServiceError{Code: "REDIS_ERROR"}
		}
		if exists == 0 {
			return GetShortenedURLOutputByBackhalf{}, &commons.ServiceError{Code: "BACKHALF_NOT_EXISTS"}
		}
	}
	// get data
	{
		err = redis_client.Get(input_request_context, commons.BH_PREFIX+input_backhalf).Scan(&existing_data)
		if err != nil {
			fmt.Println(err)
			return GetShortenedURLOutputByBackhalf{}, &commons.ServiceError{Code: "REDIS_ERROR"}
		}
	}

	if existing_data.Password != nil {
		if input_requester_id != nil && *input_requester_id == *existing_data.Creator_id {
			return GetShortenedURLOutputByBackhalf(commons.TypeReturnedURLInfo{
				Original_url:       existing_data.Original_url,
				Backhalf:           input_backhalf,
				Password_protected: true,
			}), nil
		}
		if input_password == nil {
			return GetShortenedURLOutputByBackhalf{}, &commons.ServiceError{Code: "PASSWORD_REQUIRED"}
		}
		if commons.HashString(*input_password) == *existing_data.Password {
			return GetShortenedURLOutputByBackhalf(commons.TypeReturnedURLInfo{
				Original_url:       existing_data.Original_url,
				Backhalf:           input_backhalf,
				Password_protected: true,
			}), nil
		} else {
			return GetShortenedURLOutputByBackhalf{}, &commons.ServiceError{Code: "WRONG_PASSWORD"}
		}
	} else {
		return GetShortenedURLOutputByBackhalf(commons.TypeReturnedURLInfo{
			Original_url:       existing_data.Original_url,
			Backhalf:           input_backhalf,
			Password_protected: false,
		}), nil
	}
}

type GetShortenedURLOutputByUser = struct {
	total int64
	data  []commons.TypeReturnedURLInfo
}

func (svc *URLService) getShortenedURLByUser(
	input_request_context context.Context,
	input_requester_id string,
	input_user_id string,
	input_page int,
	input_per_page int64,
) (GetShortenedURLOutputByUser, *commons.ServiceError) {
	// variable declaration
	var redis_client = commons.GetRedisClient()
	var err error
	var backhalf_list []string
	var existing_data []struct {
		commons.TypeStoredURLInfo
		backhalf string
	}
	var returned_data []commons.TypeReturnedURLInfo
	var total_data int64

	// if requester_id is not the same as user_id, FORBIDDEN
	{
		if input_requester_id != input_user_id {
			return GetShortenedURLOutputByUser{}, &commons.ServiceError{Code: "SAME_USER_REQUIRED"}
		}
	}

	// page validation
	{
		if input_page < 1 {
			return GetShortenedURLOutputByUser{}, &commons.ServiceError{Code: "INVALID_PAGE"}
		}
		if input_per_page < 1 {
			return GetShortenedURLOutputByUser{}, &commons.ServiceError{Code: "INVALID_PER_PAGE"}
		}
	}

	// get total data
	{
		total_data, err = redis_client.SCard(input_request_context, commons.USER_PREFIX+input_user_id).Result()
		if err != nil {
			fmt.Println(err)
			return GetShortenedURLOutputByUser{}, &commons.ServiceError{Code: "REDIS_ERROR"}
		}
	}

	// get backhalf list
	// 1. retrieve all backhalfs
	// 2. paginate
	{
		backhalf_list, err = redis_client.SMembers(input_request_context, commons.USER_PREFIX+input_user_id).Result()
		if err != nil {
			fmt.Println(err)
			return GetShortenedURLOutputByUser{}, &commons.ServiceError{Code: "REDIS_ERROR"}
		}
		start := (input_page - 1) * int(input_per_page)
		end := start + int(input_per_page)
		if start >= len(backhalf_list) {
			return GetShortenedURLOutputByUser{
				total: total_data,
				data:  []commons.TypeReturnedURLInfo{},
			}, nil
		}
		if end > len(backhalf_list) {
			end = len(backhalf_list)
		}
		backhalf_list = backhalf_list[total_data-int64(end) : total_data-int64(start)]
		// reverse backhalf_list
		for i := 0; i < len(backhalf_list)/2; i++ {
			backhalf_list[i], backhalf_list[len(backhalf_list)-1-i] = backhalf_list[len(backhalf_list)-1-i], backhalf_list[i]
		}
	}

	// get data
	{
		for _, backhalf := range backhalf_list {
			var data struct {
				commons.TypeStoredURLInfo
				backhalf string
			}
			err = redis_client.Get(input_request_context, commons.BH_PREFIX+backhalf).Scan(&data)
			if err != nil {
				fmt.Println(err)
				return GetShortenedURLOutputByUser{}, &commons.ServiceError{Code: "REDIS_ERROR"}
			}
			data.backhalf = backhalf
			existing_data = append(existing_data, data)
		}
	}

	// construct output
	{
		for _, data := range existing_data {
			returned_data = append(returned_data, commons.TypeReturnedURLInfo{
				Original_url:       data.Original_url,
				Backhalf:           data.backhalf,
				Password_protected: data.Password != nil,
			})
		}
	}

	return GetShortenedURLOutputByUser{
		total: total_data,
		data:  returned_data,
	}, nil
}

// UPDATE

type UpdateShortenedURLOutput = commons.TypeReturnedURLInfo

func (svc *URLService) updateShortenedURL(
	input_request_context context.Context,
	input_original_backhalf string,
	input_updater_id string,
	input_changed_backhalf *string, // nil if not changed
	input_original_url *string, // nil if not changed
	input_password *string, // nil if not changed
) (UpdateShortenedURLOutput, *commons.ServiceError) {
	// variable declaration
	var redis_client = commons.GetRedisClient()
	var exists int64
	var err error
	var original_data commons.TypeStoredURLInfo

	// check if original backhalf exists
	{
		exists, err = redis_client.Exists(input_request_context, commons.BH_PREFIX+input_original_backhalf).Result()
		if err != nil {
			fmt.Println(err)
			return UpdateShortenedURLOutput{}, &commons.ServiceError{Code: "REDIS_ERROR"}
		}
		if exists == 0 {
			return UpdateShortenedURLOutput{}, &commons.ServiceError{Code: "BACKHALF_NOT_EXISTS"}
		}
	}

	// check if desired backhalf exists
	{
		if input_changed_backhalf != nil {
			exists, err = redis_client.Exists(input_request_context, commons.BH_PREFIX+(*input_changed_backhalf)).Result()
			if err != nil {
				fmt.Println(err)
				return UpdateShortenedURLOutput{}, &commons.ServiceError{Code: "REDIS_ERROR"}
			}
			if exists == 1 {
				return UpdateShortenedURLOutput{}, &commons.ServiceError{Code: "BACKHALF_EXISTS"}
			}
		}
	}

	// get original data
	{
		err = redis_client.Get(input_request_context, commons.BH_PREFIX+input_original_backhalf).Scan(&original_data)
		if err != nil {
			fmt.Println(err)
			return UpdateShortenedURLOutput{}, &commons.ServiceError{Code: "REDIS_ERROR"}
		}
	}

	// cannot update if creator_id is nil or different
	{
		if original_data.Creator_id == nil || *original_data.Creator_id != input_updater_id {
			return UpdateShortenedURLOutput{}, &commons.ServiceError{Code: "SAME_USER_REQUIRED"}
		}
	}

	// update password
	{
		if input_password != nil {
			*input_password = commons.HashString(*input_password)
			original_data.Password = input_password
		}
	}

	// update original url
	{
		if input_original_url != nil {
			original_data.Original_url = *input_original_url
		}
	}

	// write to redis with original backhalf
	{
		err = redis_client.Set(input_request_context,
			commons.BH_PREFIX+input_original_backhalf, original_data, 0,
		).Err()
		if err != nil {
			fmt.Println(err)
			return UpdateShortenedURLOutput{}, &commons.ServiceError{Code: "REDIS_ERROR"}
		}
	}

	// if changed backhalf, update backhalf
	{
		if input_changed_backhalf != nil {
			err = redis_client.Rename(input_request_context,
				commons.BH_PREFIX+input_original_backhalf, commons.BH_PREFIX+(*input_changed_backhalf),
			).Err()
			if err != nil {
				fmt.Println(err)
				return UpdateShortenedURLOutput{}, &commons.ServiceError{Code: "REDIS_ERROR"}
			}
			// also update user data
			if original_data.Creator_id != nil {
				err = redis_client.SRem(input_request_context,
					commons.USER_PREFIX+(*original_data.Creator_id), input_original_backhalf,
				).Err()
				if err != nil {
					fmt.Println(err)
					return UpdateShortenedURLOutput{}, &commons.ServiceError{Code: "REDIS_ERROR"}
				}
				err = redis_client.SAdd(input_request_context,
					commons.USER_PREFIX+(*original_data.Creator_id), *input_changed_backhalf,
				).Err()
				if err != nil {
					fmt.Println(err)
					return UpdateShortenedURLOutput{}, &commons.ServiceError{Code: "REDIS_ERROR"}
				}
			}
		}
	}

	// construct output
	{
		returned_backhalf := input_original_backhalf
		if input_changed_backhalf != nil {
			returned_backhalf = *input_changed_backhalf
		}
		return UpdateShortenedURLOutput(commons.TypeReturnedURLInfo{
			Original_url:       original_data.Original_url,
			Backhalf:           returned_backhalf,
			Password_protected: original_data.Password != nil,
		}), nil
	}
}

type DisablePasswordProtectionOutput = commons.TypeReturnedURLInfo

func (svc *URLService) disablePasswordProtection(
	input_request_context context.Context,
	input_backhalf string,
	input_updater_id string,
) (DisablePasswordProtectionOutput, *commons.ServiceError) {
	// variable declaration
	var redis_client = commons.GetRedisClient()
	var exists int64
	var err error
	var original_data commons.TypeStoredURLInfo

	// check if backhalf exists
	{
		exists, err = redis_client.Exists(input_request_context, commons.BH_PREFIX+input_backhalf).Result()
		if err != nil {
			fmt.Println(err)
			return DisablePasswordProtectionOutput{}, &commons.ServiceError{Code: "REDIS_ERROR"}
		}
		if exists == 0 {
			return DisablePasswordProtectionOutput{}, &commons.ServiceError{Code: "BACKHALF_NOT_EXISTS"}
		}
	}

	// get original data
	{
		err = redis_client.Get(input_request_context, commons.BH_PREFIX+input_backhalf).Scan(&original_data)
		if err != nil {
			fmt.Println(err)
			return DisablePasswordProtectionOutput{}, &commons.ServiceError{Code: "REDIS_ERROR"}
		}
	}

	// cannot update if creator_id is nil or different
	{
		if original_data.Creator_id == nil || *original_data.Creator_id != input_updater_id {
			return DisablePasswordProtectionOutput{}, &commons.ServiceError{Code: "SAME_USER_REQUIRED"}
		}
	}

	// disable password protection
	{
		original_data.Password = nil
		// write to redis
		err = redis_client.Set(input_request_context, commons.BH_PREFIX+input_backhalf, original_data, 0).Err()
		if err != nil {
			fmt.Println(err)
			return DisablePasswordProtectionOutput{}, &commons.ServiceError{Code: "REDIS_ERROR"}
		}
	}

	return DisablePasswordProtectionOutput(commons.TypeReturnedURLInfo{
		Original_url:       original_data.Original_url,
		Backhalf:           input_backhalf,
		Password_protected: false,
	}), nil
}

// DELETE

type DeleteShortenedURLOutput = commons.TypeReturnedURLInfo

func (svc *URLService) deleteShortenedURL(
	input_request_context context.Context,
	input_backhalf string,
	input_deletor_id string,
) (DeleteShortenedURLOutput, *commons.ServiceError) {
	var redis_client = commons.GetRedisClient()
	var exists int64
	var err error
	var original_data commons.TypeStoredURLInfo

	// check if backhalf exists
	{
		exists, err = redis_client.Exists(input_request_context, commons.BH_PREFIX+input_backhalf).Result()
		if err != nil {
			fmt.Println(err)
			return DeleteShortenedURLOutput{}, &commons.ServiceError{Code: "REDIS_ERROR"}
		}
		if exists == 0 {
			return DeleteShortenedURLOutput{}, &commons.ServiceError{Code: "BACKHALF_NOT_EXISTS"}
		}
	}

	// get original data
	{
		err = redis_client.Get(input_request_context, commons.BH_PREFIX+input_backhalf).Scan(&original_data)
		if err != nil {
			fmt.Println(err)
			return DeleteShortenedURLOutput{}, &commons.ServiceError{Code: "REDIS_ERROR"}
		}
	}

	// cannot delete if creator_id is nil or different
	{
		if original_data.Creator_id == nil || *original_data.Creator_id != input_deletor_id {
			return DeleteShortenedURLOutput{}, &commons.ServiceError{Code: "SAME_USER_REQUIRED"}
		}
	}

	// delete from backhalf data
	{
		err = redis_client.Del(input_request_context, commons.BH_PREFIX+input_backhalf).Err()
		if err != nil {
			fmt.Println(err)
			return DeleteShortenedURLOutput{}, &commons.ServiceError{Code: "REDIS_ERROR"}
		}
	}

	// delete from user data
	{
		if original_data.Creator_id != nil {
			err = redis_client.SRem(input_request_context, commons.USER_PREFIX+(*original_data.Creator_id), input_backhalf).Err()
			if err != nil {
				fmt.Println(err)
				return DeleteShortenedURLOutput{}, &commons.ServiceError{Code: "REDIS_ERROR"}
			}
		}
	}

	return DeleteShortenedURLOutput(commons.TypeReturnedURLInfo{
		Original_url:       original_data.Original_url,
		Backhalf:           input_backhalf,
		Password_protected: original_data.Password != nil,
	}), nil
}
