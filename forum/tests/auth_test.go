package integration

// import (
//     "context"
//     "testing"
//     "time"

//     pb "back/proto"
//     "google.golang.org/grpc"
// )

// func TestRegisterUser(t *testing.T) {
//     conn, err := grpc.Dial("localhost:50051", grpc.WithInsecure())
//     if err != nil {
//         t.Fatalf("Failed to connect to auth_service: %v", err)
//     }
//     defer conn.Close()

//     client := pb.NewAuthServiceClient(conn)

//     ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
//     defer cancel()

//     req := &pb.RegisterRequest{
//         Username: "testuser",
//         Password: "testpass",
//     }

//     res, err := client.Register(ctx, req)
//     if err != nil {
//         t.Fatalf("Register failed: %v", err)
//     }

//     if res.AccessToken == "" {
//         t.Errorf("Expected access token, got empty string")
//     }
// }

// func TestAuthService_RegisterLoginFlow(t *testing.T) {
// 	conn, err := grpc.Dial(authServiceAddress, grpc.WithInsecure())
// 	require.NoError(t, err)
// 	defer conn.Close()

// 	client := proto.NewAuthServiceClient(conn)

// 	username := "testuser"
// 	password := "testpass"

// 	// Register
// 	regResp, err := client.Register(context.TODO(), &proto.RegisterRequest{
// 		Username: username,
// 		Password: password,
// 	})
// 	require.NoError(t, err)
// 	require.NotEmpty(t, regResp.AccessToken)
// 	require.NotEmpty(t, regResp.RefreshToken)

// 	// Login
// 	loginResp, err := client.Login(context.TODO(), &proto.LoginRequest{
// 		Username: username,
// 		Password: password,
// 	})
// 	require.NoError(t, err)
// 	require.NotEmpty(t, loginResp.AccessToken)
// 	require.Equal(t, username, loginResp.User.Username)
// }

// func TestAuthService_TokenValidationAndLogout(t *testing.T) {
// 	// После успешного логина
// 	token := "полученный_access_token"

// 	client := proto.NewAuthServiceClient(conn)

// 	// ValidateToken
// 	validateResp, err := client.ValidateToken(ctx, &proto.ValidateRequest{
// 		AccessToken: token,
// 	})
// 	require.NoError(t, err)
// 	require.True(t, validateResp.IsValid)

// 	// Logout
// 	logoutResp, err := client.Logout(ctx, &proto.LogoutRequest{
// 		AccessToken: token,
// 	})
// 	require.NoError(t, err)
// 	require.True(t, logoutResp.Success)
// }
