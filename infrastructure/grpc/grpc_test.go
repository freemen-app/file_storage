package grpcApi_test

import (
	"bufio"
	"context"
	"errors"
	"io"
	"os"
	"testing"
	"time"

	"github.com/alecthomas/units"
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/freemen-app/file_storage/config"
	"github.com/freemen-app/file_storage/domain/dto"
	"github.com/freemen-app/file_storage/infrastructure/app"
	grpcApi "github.com/freemen-app/file_storage/infrastructure/grpc"
	pb "github.com/freemen-app/file_storage/infrastructure/proto"
	"github.com/freemen-app/file_storage/infrastructure/testing/helpers"
	"github.com/freemen-app/file_storage/infrastructure/testing/mocks"
	fileUseCase "github.com/freemen-app/file_storage/usecase/file"
)

var (
	conf        *config.Config
	application *app.App
)

func TestMain(m *testing.M) {
	conf = config.New(config.DefaultConfig)
	application = app.New(conf)
	code := m.Run()
	os.Exit(code)
}

func testServer(t *testing.T, config *config.ApiConfig) *grpcApi.API {
	t.Helper()
	var api *grpcApi.API
	assert.NotPanics(t, func() {
		api = grpcApi.New(application, config)
		go api.Start()
	})
	t.Cleanup(api.Shutdown)
	return api
}

func testClient(t *testing.T, config *config.ApiConfig) pb.FileStorageClient {
	t.Helper()
	conn, err := grpc.DialContext(
		helpers.TimeoutCtx(t, helpers.DefaultCtx, time.Second),
		config.Addr(),
		grpc.WithInsecure(),
	)
	assert.NoError(t, err)
	client := pb.NewFileStorageClient(conn)
	t.Cleanup(func() {
		assert.NoError(t, conn.Close())
	})
	return client
}

func TestNew(t *testing.T) {
	tests := []struct {
		name      string
		port      int
		wantPanic bool
	}{
		{
			name: "succeed",
			port: 9999,
		},
		{
			name:      "port in use",
			port:      9999,
			wantPanic: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			conf := &config.ApiConfig{Host: "localhost", Port: tt.port}
			test := func() {
				api := grpcApi.New(application, conf)
				assert.NotNil(t, api.Grpc())
				assert.NotNil(t, api.Handler())
				assert.NotNil(t, api.Listener())
			}
			if tt.wantPanic {
				assert.Panics(t, test)
			} else {
				assert.NotPanics(t, test)
			}
		})
	}
}

func TestGrpcApi_Start_Shutdown(t *testing.T) {
	apiConf := &config.ApiConfig{Host: "localhost", Port: 9090}
	api := grpcApi.New(application, apiConf)
	serverRunning := make(chan bool)
	go func() {
		serverRunning <- true
		api.Start()
		serverRunning <- false
	}()
	assert.True(t, <-serverRunning)
	time.Sleep(time.Second / 4)
	api.Shutdown()
	assert.False(t, <-serverRunning)
}

func TestNewHandler(t *testing.T) {
	wantUseCase := fileUseCase.New(new(mocks.FileRepo))
	h := grpcApi.NewHandler(wantUseCase)
	assert.EqualValues(t, wantUseCase, h.FileUseCase())
	assert.NotNil(t, h.Presenter())
}

func TestHandler_Upload(t *testing.T) {
	conf := &config.ApiConfig{Host: "localhost", Port: 9998}
	server := testServer(t, conf)
	client := testClient(t, conf)

	type args struct {
		ctx      context.Context
		metadata *pb.MetaData
		file     io.Reader
	}
	tests := []struct {
		name        string
		args        args
		mockCalls   helpers.MockCalls
		wantUrl     string
		wantErrCode codes.Code
	}{
		{
			name: "succeed 1mb",
			args: args{
				ctx:      helpers.DefaultCtx,
				metadata: &pb.MetaData{Directory: "test", Filename: "1mb.jpg"},
				file:     helpers.OpenFile(t, "testdata/1mb.jpg"),
			},
			mockCalls: helpers.MockCalls{
				{
					Method:     "Upload",
					Args:       []interface{}{mock.Anything, mock.Anything},
					ReturnArgs: []interface{}{"https://aws.s3/test/1mb.jpg", nil},
				},
			},
			wantUrl:     "https://aws.s3/test/1mb.jpg",
			wantErrCode: codes.OK,
		},
		{
			name: "succeed 5mb",
			args: args{
				ctx:      helpers.DefaultCtx,
				metadata: &pb.MetaData{Directory: "test", Filename: "5mb.png"},
				file:     helpers.OpenFile(t, "testdata/5mb.png"),
			},
			mockCalls: helpers.MockCalls{
				{
					Method:     "Upload",
					Args:       []interface{}{mock.Anything, mock.Anything},
					ReturnArgs: []interface{}{"https://aws.s3/test/5mb.png", nil},
				},
			},
			wantUrl:     "https://aws.s3/test/5mb.png",
			wantErrCode: codes.OK,
		},
		{
			name: "timeout",
			args: args{
				ctx:      helpers.TimeoutCtx(t, helpers.DefaultCtx, time.Nanosecond),
				metadata: &pb.MetaData{Directory: "test", Filename: "test.jpg"},
				file:     helpers.OpenFile(t, "testdata/1mb.jpg"),
			},
			wantErrCode: codes.DeadlineExceeded,
		},
		{
			name: "validation error",
			args: args{
				ctx:      helpers.DefaultCtx,
				metadata: &pb.MetaData{Directory: "test", Filename: "test.jpg"},
				file:     helpers.OpenFile(t, "testdata/1mb.jpg"),
			},
			mockCalls: helpers.MockCalls{
				{
					Method:     "Upload",
					Args:       []interface{}{mock.Anything, mock.Anything},
					ReturnArgs: []interface{}{"", validation.Errors{}},
				},
			},
			wantErrCode: codes.InvalidArgument,
		},
		{
			name: "validation error",
			args: args{
				ctx:      helpers.DefaultCtx,
				metadata: &pb.MetaData{Directory: "test", Filename: "test.jpg"},
				file:     helpers.OpenFile(t, "testdata/1mb.jpg"),
			},
			mockCalls: helpers.MockCalls{
				{
					Method:     "Upload",
					Args:       []interface{}{mock.Anything, mock.Anything},
					ReturnArgs: []interface{}{"", errors.New("test error")},
				},
			},
			wantErrCode: codes.Internal,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			useCase := new(mocks.FileUseCase)
			for _, call := range tt.mockCalls {
				useCase.On(call.Method, call.Args...).Return(call.ReturnArgs...)
			}
			server.Handler().SetFileUseCase(useCase)

			test := func() (string, error) {
				stream, err := client.Upload(tt.args.ctx)
				if err != nil {
					return "", err
				}

				uploadRequest := &pb.UploadRequest{File: &pb.UploadRequest_Metadata{Metadata: tt.args.metadata}}
				if err := stream.Send(uploadRequest); err != nil {
					return "", nil
				}

				// Send file by chunks
				reader := bufio.NewReader(tt.args.file)
				chunk := make([]byte, 0, 500*units.KB)
				for {
					n, err := reader.Read(chunk[:cap(chunk)])
					if err == io.EOF {
						break
					} else if err != nil {
						assert.NoError(t, stream.CloseSend())
						return "", err
					}
					request := &pb.UploadRequest{
						File: &pb.UploadRequest_Content{Content: chunk[:n]},
					}
					if err := stream.Send(request); err != nil {
						return "", err
					}
				}
				response, err := stream.CloseAndRecv()
				if err != nil {
					return "", err
				}
				return response.Url, nil
			}

			gotUrl, gotErr := test()
			grpcErr, ok := status.FromError(gotErr)
			assert.True(t, ok)
			assert.EqualValues(t, tt.wantErrCode, grpcErr.Code(), grpcErr.Message())
			assert.EqualValues(t, tt.wantUrl, gotUrl)

			useCase.AssertExpectations(t)
		})
	}
}

func TestHandler_Delete(t *testing.T) {
	conf := &config.ApiConfig{Host: "localhost", Port: 9998}
	server := testServer(t, conf)
	client := testClient(t, conf)

	type args struct {
		ctx context.Context
		in  *pb.DeleteRequest
	}
	tests := []struct {
		name        string
		args        args
		mockCalls   helpers.MockCalls
		wantErrCode codes.Code
	}{
		{
			name: "succeed",
			args: args{
				ctx: helpers.DefaultCtx,
				in:  &pb.DeleteRequest{Url: "https://aws.s3/bucket/test.jpg"},
			},
			mockCalls: helpers.MockCalls{
				{
					Method:     "Delete",
					Args:       []interface{}{mock.Anything, dto.DeleteInput("https://aws.s3/bucket/test.jpg")},
					ReturnArgs: []interface{}{nil},
				},
			},
			wantErrCode: codes.OK,
		},
		{
			name: "timeout",
			args: args{
				ctx: helpers.TimeoutCtx(t, helpers.DefaultCtx, time.Nanosecond),
				in:  &pb.DeleteRequest{Url: "https://aws.s3/bucket/test.jpg"},
			},
			wantErrCode: codes.DeadlineExceeded,
		},
		{
			name: "validation error",
			args: args{
				ctx: helpers.DefaultCtx,
				in:  &pb.DeleteRequest{Url: "https://aws.s3/bucket/test.jpg"},
			},
			mockCalls: helpers.MockCalls{
				{
					Method:     "Delete",
					Args:       []interface{}{mock.Anything, dto.DeleteInput("https://aws.s3/bucket/test.jpg")},
					ReturnArgs: []interface{}{validation.Errors{}},
				},
			},
			wantErrCode: codes.InvalidArgument,
		},
		{
			name: "internal error",
			args: args{
				ctx: helpers.DefaultCtx,
				in:  &pb.DeleteRequest{Url: "https://aws.s3/bucket/test.jpg"},
			},
			mockCalls: helpers.MockCalls{
				{
					Method:     "Delete",
					Args:       []interface{}{mock.Anything, dto.DeleteInput("https://aws.s3/bucket/test.jpg")},
					ReturnArgs: []interface{}{errors.New("test error")},
				},
			},
			wantErrCode: codes.Internal,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			useCase := new(mocks.FileUseCase)
			for _, call := range tt.mockCalls {
				useCase.On(call.Method, call.Args...).Return(call.ReturnArgs...)
			}
			server.Handler().SetFileUseCase(useCase)

			_, gotErr := client.Delete(tt.args.ctx, tt.args.in)
			grpcErr, ok := status.FromError(gotErr)
			assert.True(t, ok)
			assert.EqualValues(t, tt.wantErrCode, grpcErr.Code(), grpcErr.Message())

			useCase.AssertExpectations(t)
		})
	}
}

func TestHandler_BatchDelete(t *testing.T) {
	conf := &config.ApiConfig{Host: "localhost", Port: 9998}
	server := testServer(t, conf)
	client := testClient(t, conf)

	type args struct {
		ctx context.Context
		in  *pb.BatchDeleteRequest
	}
	tests := []struct {
		name        string
		args        args
		mockCalls   helpers.MockCalls
		wantErrCode codes.Code
	}{
		{
			name: "succeed",
			args: args{
				ctx: helpers.DefaultCtx,
				in: &pb.BatchDeleteRequest{Urls: []string{
					"https://aws.s3/bucket/test.jpg",
					"https://aws.s3/bucket/test2.jpg",
				}},
			},
			mockCalls: helpers.MockCalls{
				{
					Method: "BatchDelete",
					Args: []interface{}{
						mock.Anything,
						dto.BatchDeleteInput{
							"https://aws.s3/bucket/test.jpg",
							"https://aws.s3/bucket/test2.jpg",
						},
					},
					ReturnArgs: []interface{}{nil},
				},
			},
			wantErrCode: codes.OK,
		},
		{
			name: "timeout",
			args: args{
				ctx: helpers.TimeoutCtx(t, helpers.DefaultCtx, time.Nanosecond),
				in: &pb.BatchDeleteRequest{Urls: []string{
					"https://aws.s3/bucket/test.jpg",
					"https://aws.s3/bucket/test2.jpg",
				}},
			},
			wantErrCode: codes.DeadlineExceeded,
		},
		{
			name: "validation error",
			args: args{
				ctx: helpers.DefaultCtx,
				in: &pb.BatchDeleteRequest{Urls: []string{
					"https://aws.s3/bucket/test.jpg",
					"https://aws.s3/bucket/test2.jpg",
				}},
			},
			mockCalls: helpers.MockCalls{
				{
					Method: "BatchDelete",
					Args: []interface{}{
						mock.Anything,
						dto.BatchDeleteInput{
							"https://aws.s3/bucket/test.jpg",
							"https://aws.s3/bucket/test2.jpg",
						},
					},
					ReturnArgs: []interface{}{validation.Errors{}},
				},
			},
			wantErrCode: codes.InvalidArgument,
		},
		{
			name: "internal error",
			args: args{
				ctx: helpers.DefaultCtx,
				in: &pb.BatchDeleteRequest{Urls: []string{
					"https://aws.s3/bucket/test.jpg",
					"https://aws.s3/bucket/test2.jpg",
				}},
			},
			mockCalls: helpers.MockCalls{
				{
					Method: "BatchDelete",
					Args: []interface{}{
						mock.Anything,
						dto.BatchDeleteInput{
							"https://aws.s3/bucket/test.jpg",
							"https://aws.s3/bucket/test2.jpg",
						},
					},
					ReturnArgs: []interface{}{errors.New("test error")},
				},
			},
			wantErrCode: codes.Internal,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			useCase := new(mocks.FileUseCase)
			for _, call := range tt.mockCalls {
				useCase.On(call.Method, call.Args...).Return(call.ReturnArgs...)
			}
			server.Handler().SetFileUseCase(useCase)

			_, gotErr := client.BatchDelete(tt.args.ctx, tt.args.in)
			grpcErr, ok := status.FromError(gotErr)
			assert.True(t, ok)
			assert.EqualValues(t, tt.wantErrCode, grpcErr.Code(), grpcErr.Message())

			useCase.AssertExpectations(t)
		})
	}
}

// func TestHandler_Get(t *testing.T) {
// 	conf := &config.ApiConfig{Host: "localhost", Port: 9997}
// 	server := testServer(t, conf, application)
// 	client := testClient(t, conf)
//
// 	type args struct {
// 		ctx   context.Context
// 		input *pb.GetByIdRequest
// 	}
// 	type returnArgs struct {
// 		user *entity.User
// 		err  error
// 	}
// 	tests := []struct {
// 		name        string
// 		args        args
// 		returnArgs  returnArgs
// 		wantUser    *pb.User
// 		wantErrCode codes.Code
// 	}{
// 		{
// 			name: "succeed",
// 			args: args{
// 				ctx:   helpers.DefaultCtx,
// 				input: &pb.GetByIdRequest{Id: "5f56383eff0f482241877980"},
// 			},
// 			returnArgs: returnArgs{
// 				user: &entity.User{
// 					ID:        "5f56383eff0f482241877980",
// 					Username:  "test",
// 					Email:     "test.user@gmail.com",
// 					BirthDate: &civil.Date{Year: 2020, Month: 12, Day: 6},
// 				},
// 			},
// 			wantUser: &pb.User{
// 				Id:       "5f56383eff0f482241877980",
// 				Username: "test",
// 				Email:    "test.user@gmail.com",
// 				BirthDate: &timestamp.Timestamp{
// 					Seconds: civil.Date{Year: 2020, Month: 12, Day: 6}.In(time.UTC).Unix(),
// 				},
// 			},
// 		},
// 		{
// 			name: "internal error",
// 			args: args{
// 				ctx:   helpers.DefaultCtx,
// 				input: &pb.GetByIdRequest{Id: "12345"},
// 			},
// 			returnArgs:  returnArgs{err: errors.New("test")},
// 			wantErrCode: codes.Internal,
// 		},
// 		{
// 			name: "not found error",
// 			args: args{
// 				ctx:   helpers.DefaultCtx,
// 				input: &pb.GetByIdRequest{Id: "12345"},
// 			},
// 			returnArgs:  returnArgs{err: customErrors.ErrUserNotFound},
// 			wantErrCode: codes.NotFound,
// 		},
// 		{
// 			name: "validation error",
// 			args: args{
// 				ctx:   helpers.DefaultCtx,
// 				input: &pb.GetByIdRequest{Id: "12345"},
// 			},
// 			returnArgs:  returnArgs{err: validation.Errors{}},
// 			wantErrCode: codes.InvalidArgument,
// 		},
// 		{
// 			name: "timeout error",
// 			args: args{
// 				ctx:   helpers.DefaultCtx,
// 				input: &pb.GetByIdRequest{Id: "12345"},
// 			},
// 			returnArgs:  returnArgs{err: context.DeadlineExceeded},
// 			wantErrCode: codes.DeadlineExceeded,
// 		},
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			useCase := new(mocks.UserUseCase)
// 			useCase.On(
// 				"Get", mock.Anything, mock.Anything,
// 			).Return(tt.returnArgs.user, tt.returnArgs.err)
// 			server.Handler().SetUserUseCase(useCase)
//
// 			gotUser, gotErr := client.Get(tt.args.ctx, tt.args.input)
// 			grpcErr, ok := status.FromError(gotErr)
// 			assert.True(t, ok)
// 			assert.EqualValues(t, tt.wantErrCode, grpcErr.Code(), grpcErr.Message())
// 			equalUsers(t, tt.wantUser, gotUser)
// 		})
// 	}
// }
//
// func TestHandler_Update(t *testing.T) {
// 	conf := &config.ApiConfig{Host: "localhost", Port: 9996}
// 	server := testServer(t, conf, application)
// 	client := testClient(t, conf)
//
// 	type args struct {
// 		ctx   context.Context
// 		input *pb.UpdateRequest
// 	}
// 	type returnArgs struct {
// 		user *entity.User
// 		err  error
// 	}
// 	tests := []struct {
// 		name        string
// 		args        args
// 		returnArgs  returnArgs
// 		wantUser    *pb.User
// 		wantErrCode codes.Code
// 	}{
// 		{
// 			name: "succeed",
// 			args: args{
// 				ctx:   helpers.DefaultCtx,
// 				input: &pb.UpdateRequest{},
// 			},
// 			returnArgs: returnArgs{
// 				user: &entity.User{
// 					ID:        "5f56383eff0f482241877980",
// 					Username:  "test",
// 					Email:     "test.user@gmail.com",
// 					BirthDate: &civil.Date{Year: 2020, Month: 12, Day: 6},
// 				},
// 			},
// 			wantUser: &pb.User{
// 				Id:       "5f56383eff0f482241877980",
// 				Username: "test",
// 				Email:    "test.user@gmail.com",
// 				BirthDate: &timestamp.Timestamp{
// 					Seconds: civil.Date{Year: 2020, Month: 12, Day: 6}.In(time.UTC).Unix(),
// 				},
// 			},
// 		},
// 		{
// 			name: "timeout error",
// 			args: args{
// 				ctx:   helpers.DefaultCtx,
// 				input: &pb.UpdateRequest{Id: "5f56383eff0f482241877980"},
// 			},
// 			returnArgs:  returnArgs{err: context.DeadlineExceeded},
// 			wantErrCode: codes.DeadlineExceeded,
// 		},
// 		{
// 			name: "validation error",
// 			args: args{
// 				ctx:   helpers.DefaultCtx,
// 				input: &pb.UpdateRequest{Id: "5f56383eff0f482241877980"},
// 			},
// 			returnArgs:  returnArgs{err: validation.Errors{}},
// 			wantErrCode: codes.InvalidArgument,
// 		},
// 		{
// 			name: "not found",
// 			args: args{
// 				ctx:   helpers.DefaultCtx,
// 				input: &pb.UpdateRequest{Id: "5f56383eff0f482241877980"},
// 			},
// 			returnArgs:  returnArgs{err: customErrors.ErrUserNotFound},
// 			wantErrCode: codes.NotFound,
// 		},
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			useCase := new(mocks.UserUseCase)
// 			useCase.On(
// 				"Update", mock.Anything, mock.Anything,
// 			).Return(tt.returnArgs.user, tt.returnArgs.err)
// 			server.Handler().SetUserUseCase(useCase)
//
// 			gotUser, gotErr := client.Update(tt.args.ctx, tt.args.input)
// 			grpcErr, ok := status.FromError(gotErr)
// 			assert.True(t, ok)
// 			assert.EqualValues(t, tt.wantErrCode, grpcErr.Code(), grpcErr.Message())
// 			equalUsers(t, tt.wantUser, gotUser)
//
// 			useCase.AssertExpectations(t)
// 		})
// 	}
// }
//
// //
// // func TestHandler_UpdateAvatar(t *testing.T) {
// // 	conf := &config.ApiConfig{Host: "localhost", Port: 9995}
// // 	server := testServer(t, conf, application)
// // 	client := testClient(t, conf)
// //
// // 	type args struct {
// // 		ctx      context.Context
// // 		metadata *pb.MetaData
// // 		file     *os.File
// // 	}
// // 	tests := []struct {
// // 		name        string
// // 		userUseCase userUseCase.UseCase
// // 		args        args
// // 		wantUser    *pb.User
// // 		wantErrCode codes.Code
// // 	}{
// // 		{
// // 			name: "succeed",
// // 			userUseCase: &mocks.UserUseCase{wantReturn: &entity.User{
// // 				ID:       "5f56383eff0f482241877980",
// // 				Username: "test",
// // 				Email:    "test.email@gmail.com",
// // 				Avatar:   "https://aws.s3/test/test.jpg",
// // 			}, shouldValidate: true},
// // 			args: args{
// // 				ctx: helpers.DefaultCtx,
// // 				metadata: &pb.MetaData{
// // 					UserId:   "5f56383eff0f482241877980",
// // 					Filename: "test.jpg",
// // 				},
// // 				file: helpers.OpenFile(t, "grpc_test.go"),
// // 			},
// // 			wantUser: &pb.User{
// // 				Id:       "5f56383eff0f482241877980",
// // 				Username: "test",
// // 				Email:    "test.email@gmail.com",
// // 				Avatar:   "https://aws.s3/test/test.jpg",
// // 			},
// // 		},
// // 		{
// // 			name:        "timeout error",
// // 			userUseCase: &mocks.UserUseCase{},
// // 			args: args{
// // 				ctx: helpers.TimeoutCtx(t, helpers.DefaultCtx, 0),
// // 				metadata: &pb.MetaData{
// // 					UserId:   "5f56383eff0f482241877980",
// // 					Filename: "test.jpg",
// // 				},
// // 				file: helpers.OpenFile(t, "grpc_test.go"),
// // 			},
// // 			wantErrCode: codes.DeadlineExceeded,
// // 		},
// // 		{
// // 			name:        "validation error",
// // 			userUseCase: &mocks.UserUseCase{shouldValidate: true},
// // 			args: args{
// // 				ctx: helpers.DefaultCtx,
// // 				metadata: &pb.MetaData{
// // 					UserId:   "5f56383eff0f482241877980",
// // 					Filename: "test.go",
// // 				},
// // 				file: helpers.OpenFile(t, "grpc_test.go"),
// // 			},
// // 			wantErrCode: codes.InvalidArgument,
// // 		},
// // 	}
// // 	for _, tt := range tests {
// // 		t.Run(tt.name, func(t *testing.T) {
// // 			server.Handler().SetUserUseCase(tt.userUseCase)
// // 			test := func() (*pb.User, error) {
// // 				upload, err := client.UpdateAvatar(tt.args.ctx)
// // 				if err != nil {
// // 					return nil, err
// // 				}
// // 				// Send Metadata
// // 				request := &pb.UpdateAvatarRequest{
// // 					Avatar: &pb.UpdateAvatarRequest_Metadata{Metadata: tt.args.metadata},
// // 				}
// // 				if err := upload.Send(request); err != nil {
// // 					return nil, err
// // 				}
// // 				// Send File by chunks of 1024 bytes
// // 				reader := bufio.NewReader(tt.args.file)
// // 				chunk := make([]byte, 0, 1024)
// // 				for {
// // 					n, err := reader.Read(chunk[:cap(chunk)])
// // 					if err == io.EOF {
// // 						break
// // 					} else if err != nil {
// // 						assert.NoError(t, upload.CloseSend())
// // 						return nil, err
// // 					}
// // 					request := &pb.UpdateAvatarRequest{
// // 						Avatar: &pb.UpdateAvatarRequest_Content{Content: chunk[:n]},
// // 					}
// // 					if err := upload.Send(request); err != nil {
// // 						return nil, err
// // 					}
// // 				}
// // 				return upload.CloseAndRecv()
// // 			}
// //
// // 			gotUser, gotErr := test()
// // 			grpcErr, ok := status.FromError(gotErr)
// // 			assert.True(t, ok)
// // 			assert.EqualValues(t, tt.wantErrCode, grpcErr.Code(), grpcErr.Message())
// // 			equalUsers(t, tt.wantUser, gotUser)
// // 		})
// // 	}
// // }
//
// func TestHandler_DelayedDelete(t *testing.T) {
// 	conf := &config.ApiConfig{Host: "localhost", Port: 9993}
// 	server := testServer(t, conf, application)
// 	client := testClient(t, conf)
//
// 	type args struct {
// 		ctx   context.Context
// 		input *pb.GetByIdRequest
// 	}
// 	type returnArgs struct {
// 		err error
// 	}
// 	tests := []struct {
// 		name        string
// 		args        args
// 		returnArgs  returnArgs
// 		wantErrCode codes.Code
// 	}{
// 		{
// 			name: "succeed",
// 			args: args{
// 				ctx:   helpers.DefaultCtx,
// 				input: &pb.GetByIdRequest{Id: "5f56383eff0f482241877980"},
// 			},
// 		},
// 		{
// 			name: "timeout error",
// 			args: args{
// 				ctx:   helpers.DefaultCtx,
// 				input: &pb.GetByIdRequest{Id: "5f56383eff0f482241877980"},
// 			},
// 			returnArgs:  returnArgs{err: context.DeadlineExceeded},
// 			wantErrCode: codes.DeadlineExceeded,
// 		},
// 		{
// 			name: "validation error",
// 			args: args{
// 				ctx:   helpers.DefaultCtx,
// 				input: &pb.GetByIdRequest{Id: "5f56383eff0f482241877980"},
// 			},
// 			returnArgs:  returnArgs{err: validation.Errors{}},
// 			wantErrCode: codes.InvalidArgument,
// 		},
// 		{
// 			name: "not found",
// 			args: args{
// 				ctx:   helpers.DefaultCtx,
// 				input: &pb.GetByIdRequest{Id: "5f56383eff0f482241877980"},
// 			},
// 			returnArgs:  returnArgs{err: customErrors.ErrUserNotFound},
// 			wantErrCode: codes.NotFound,
// 		},
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			useCase := new(mocks.UserUseCase)
// 			useCase.On("DelayedDelete", mock.Anything, mock.Anything).Return(tt.returnArgs.err)
// 			server.Handler().SetUserUseCase(useCase)
//
// 			_, gotErr := client.DelayedDelete(tt.args.ctx, tt.args.input)
// 			grpcErr, ok := status.FromError(gotErr)
// 			assert.True(t, ok)
// 			assert.EqualValues(t, tt.wantErrCode, grpcErr.Code(), grpcErr.Message())
//
// 			useCase.AssertExpectations(t)
// 		})
// 	}
// }
//
// func TestHandler_CreateVerification(t *testing.T) {
// 	conf := &config.ApiConfig{Host: "localhost", Port: 9992}
// 	server := testServer(t, conf, application)
// 	client := testClient(t, conf)
//
// 	type args struct {
// 		ctx   context.Context
// 		input *pb.GetByIdRequest
// 	}
// 	type returnArgs struct {
// 		err error
// 	}
// 	tests := []struct {
// 		name        string
// 		args        args
// 		returnArgs  returnArgs
// 		wantErrCode codes.Code
// 	}{
// 		{
// 			name: "succeed",
// 			args: args{
// 				ctx:   helpers.DefaultCtx,
// 				input: &pb.GetByIdRequest{Id: "5f56383eff0f482241877980"},
// 			},
// 		},
// 		{
// 			name: "timeout error",
// 			args: args{
// 				ctx:   helpers.DefaultCtx,
// 				input: &pb.GetByIdRequest{Id: "5f56383eff0f482241877980"},
// 			},
// 			returnArgs:  returnArgs{err: context.DeadlineExceeded},
// 			wantErrCode: codes.DeadlineExceeded,
// 		},
// 		{
// 			name: "validation error",
// 			args: args{
// 				ctx:   helpers.DefaultCtx,
// 				input: &pb.GetByIdRequest{Id: "5f56383eff0f482241877980"},
// 			},
// 			returnArgs:  returnArgs{err: validation.Errors{}},
// 			wantErrCode: codes.InvalidArgument,
// 		},
// 		{
// 			name: "not found",
// 			args: args{
// 				ctx:   helpers.DefaultCtx,
// 				input: &pb.GetByIdRequest{Id: "5f56383eff0f482241877980"},
// 			},
// 			returnArgs:  returnArgs{err: customErrors.ErrUserNotFound},
// 			wantErrCode: codes.NotFound,
// 		},
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			useCase := new(mocks.UserUseCase)
// 			useCase.On("CreateVerification", mock.Anything, mock.Anything).Return(tt.returnArgs.err)
// 			server.Handler().SetUserUseCase(useCase)
//
// 			_, gotErr := client.CreateVerification(tt.args.ctx, tt.args.input)
// 			grpcErr, ok := status.FromError(gotErr)
// 			assert.True(t, ok)
// 			assert.EqualValues(t, tt.wantErrCode, grpcErr.Code(), grpcErr.Message())
//
// 			useCase.AssertExpectations(t)
// 		})
// 	}
// }
//
// func TestHandler_VerifyEmail(t *testing.T) {
// 	conf := &config.ApiConfig{Host: "localhost", Port: 9991}
// 	server := testServer(t, conf, application)
// 	client := testClient(t, conf)
//
// 	type args struct {
// 		ctx   context.Context
// 		input *pb.VerifyEmailRequest
// 	}
// 	type returnArgs struct {
// 		err error
// 	}
// 	tests := []struct {
// 		name        string
// 		args        args
// 		returnArgs  returnArgs
// 		wantErrCode codes.Code
// 	}{
// 		{
// 			name: "succeed",
// 			args: args{
// 				ctx:   helpers.DefaultCtx,
// 				input: &pb.VerifyEmailRequest{Id: "5f56383eff0f482241877980"},
// 			},
// 		},
// 		{
// 			name: "timeout error",
// 			args: args{
// 				ctx:   helpers.DefaultCtx,
// 				input: &pb.VerifyEmailRequest{Id: "5f56383eff0f482241877980"},
// 			},
// 			returnArgs:  returnArgs{err: context.DeadlineExceeded},
// 			wantErrCode: codes.DeadlineExceeded,
// 		},
// 		{
// 			name: "validation error",
// 			args: args{
// 				ctx:   helpers.DefaultCtx,
// 				input: &pb.VerifyEmailRequest{Id: "5f56383eff0f482241877980"},
// 			},
// 			returnArgs:  returnArgs{err: validation.Errors{}},
// 			wantErrCode: codes.InvalidArgument,
// 		},
// 		{
// 			name: "not found",
// 			args: args{
// 				ctx:   helpers.DefaultCtx,
// 				input: &pb.VerifyEmailRequest{Id: "5f56383eff0f482241877980"},
// 			},
// 			returnArgs:  returnArgs{err: customErrors.ErrUserNotFound},
// 			wantErrCode: codes.NotFound,
// 		},
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			useCase := new(mocks.UserUseCase)
// 			useCase.On("VerifyEmail", mock.Anything, mock.Anything).Return(tt.returnArgs.err)
// 			server.Handler().SetUserUseCase(useCase)
//
// 			_, gotErr := client.VerifyEmail(tt.args.ctx, tt.args.input)
// 			grpcErr, ok := status.FromError(gotErr)
// 			assert.True(t, ok)
// 			assert.EqualValues(t, tt.wantErrCode, grpcErr.Code(), grpcErr.Message())
//
// 			useCase.AssertExpectations(t)
// 		})
// 	}
// }
