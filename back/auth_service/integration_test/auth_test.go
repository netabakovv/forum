package integration_test

import (
	"context"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"log"
	"os"
	"testing"
	"time"

	pb "github.com/netabakovv/forum/back/proto"
	"google.golang.org/grpc"
)

var authClient pb.AuthServiceClient

func TestMain(m *testing.M) {
	// Ждём поднятие сервисов
	time.Sleep(5 * time.Second)

	conn, err := grpc.Dial("localhost:50053", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("не удалось подключиться к gRPC: %v", err)
	}
	defer conn.Close()

	authClient = pb.NewAuthServiceClient(conn)

	os.Exit(m.Run())
}

func TestRegister_Success(t *testing.T) {
	ctx := context.Background()

	username := "testuser_" + time.Now().Format("150405")
	password := "securepass"

	resp, err := authClient.Register(ctx, &pb.RegisterRequest{
		Username: username,
		Password: password,
	})
	if err != nil {
		t.Fatalf("ошибка при регистрации: %v", err)
	}
	if resp.AccessToken == "" || resp.RefreshToken == "" {
		t.Fatal("ожидались токены, но не получены")
	}
}

func TestLogin_Success(t *testing.T) {
	ctx := context.Background()

	username := "loginuser_" + time.Now().Format("150405")
	password := "loginpass"

	_, _ = authClient.Register(ctx, &pb.RegisterRequest{
		Username: username,
		Password: password,
	})

	resp, err := authClient.Login(ctx, &pb.LoginRequest{
		Username: username,
		Password: password,
	})
	if err != nil {
		t.Fatalf("ошибка при логине: %v", err)
	}
	if resp.User == nil || resp.User.Username != username {
		t.Fatal("неверный пользователь или пользователь не получен")
	}
}

func TestRefreshToken_Success(t *testing.T) {
	ctx := context.Background()

	username := "refresh_" + time.Now().Format("150405")
	password := "refreshpass"

	regResp, _ := authClient.Register(ctx, &pb.RegisterRequest{
		Username: username,
		Password: password,
	})

	refResp, err := authClient.RefreshToken(ctx, &pb.RefreshTokenRequest{
		RefreshToken: regResp.RefreshToken,
	})
	if err != nil {
		t.Fatalf("ошибка обновления токена: %v", err)
	}
	if refResp.AccessToken == "" || refResp.RefreshToken == "" {
		t.Fatal("ожидались новые токены, но не получены")
	}
}

func TestValidateToken_Valid(t *testing.T) {
	ctx := context.Background()

	username := "valid_" + time.Now().Format("150405")
	password := "validpass"

	regResp, _ := authClient.Register(ctx, &pb.RegisterRequest{
		Username: username,
		Password: password,
	})

	valResp, err := authClient.ValidateToken(ctx, &pb.ValidateRequest{
		AccessToken: regResp.AccessToken,
	})
	if err != nil {
		t.Fatalf("ошибка валидации токена: %v", err)
	}
	if !valResp.IsValid || valResp.Username != username {
		t.Fatal("токен должен быть валиден, но не прошёл проверку")
	}
}

func TestLogout_Success(t *testing.T) {
	ctx := context.Background()

	username := "logout_" + time.Now().Format("150405")
	password := "logoutpass"

	_, _ = authClient.Register(ctx, &pb.RegisterRequest{
		Username: username,
		Password: password,
	})

	logResp, _ := authClient.Login(ctx, &pb.LoginRequest{Username: username, Password: password})

	resp, err := authClient.Logout(ctx, &pb.LogoutRequest{
		AccessToken: logResp.AccessToken,
	})
	if err != nil {
		t.Fatalf("ошибка при выходе: %v", err)
	}
	if !resp.Success {
		t.Fatal("ожидался успех при logout")
	}
}

func TestCheckAdminStatus_NotAdmin(t *testing.T) {
	ctx := context.Background()

	username := "notadmin_" + time.Now().Format("150405")
	password := "admincheck"

	regResp, _ := authClient.Register(ctx, &pb.RegisterRequest{
		Username: username,
		Password: password,
	})

	valResp, _ := authClient.ValidateToken(ctx, &pb.ValidateRequest{
		AccessToken: regResp.AccessToken,
	})

	resp, err := authClient.CheckAdminStatus(ctx, &pb.CheckAdminRequest{
		UserId: valResp.UserId,
	})
	if err != nil {
		t.Fatalf("ошибка проверки статуса админа: %v", err)
	}
	if resp.IsAdmin {
		t.Fatal("новый пользователь не должен быть админом")
	}
}

func TestRegister_EmptyUsernamePassword(t *testing.T) {
	ctx := context.Background()

	_, err := authClient.Register(ctx, &pb.RegisterRequest{
		Username: "",
		Password: "",
	})
	if err == nil {
		t.Fatal("ожидалась ошибка при пустом username и password")
	}
	st, _ := status.FromError(err)
	if st.Code() != codes.InvalidArgument {
		t.Fatalf("ожидался код InvalidArgument, получили %v", st.Code())
	}
}

func TestLogin_InvalidCredentials(t *testing.T) {
	ctx := context.Background()

	// Попытка логина с несуществующим пользователем
	_, err := authClient.Login(ctx, &pb.LoginRequest{
		Username: "nonexistentuser",
		Password: "wrongpass",
	})
	if err == nil {
		t.Fatal("ожидалась ошибка при неверных учетных данных")
	}
	st, _ := status.FromError(err)
	if st.Code() != codes.NotFound && st.Code() != codes.Unauthenticated {
		t.Fatalf("ожидался код NotFound или Unauthenticated, получили %v", st.Code())
	}
}

func TestRefreshToken_EmptyToken(t *testing.T) {
	ctx := context.Background()

	_, err := authClient.RefreshToken(ctx, &pb.RefreshTokenRequest{
		RefreshToken: "",
	})
	if err == nil {
		t.Fatal("ожидалась ошибка при пустом refresh token")
	}
	st, _ := status.FromError(err)
	if st.Code() != codes.InvalidArgument {
		t.Fatalf("ожидался код InvalidArgument, получили %v", st.Code())
	}
}

func TestRefreshToken_InvalidToken(t *testing.T) {
	ctx := context.Background()

	_, err := authClient.RefreshToken(ctx, &pb.RefreshTokenRequest{
		RefreshToken: "invalidtoken",
	})
	if err == nil {
		t.Fatal("ожидалась ошибка при неверном refresh token")
	}
	st, _ := status.FromError(err)
	if st.Code() != codes.Unauthenticated {
		t.Fatalf("ожидался код Unauthenticated, получили %v", st.Code())
	}
}

func TestValidateToken_EmptyToken(t *testing.T) {
	ctx := context.Background()

	_, err := authClient.ValidateToken(ctx, &pb.ValidateRequest{
		AccessToken: "",
	})
	if err == nil {
		t.Fatal("ожидалась ошибка при пустом access token")
	}
	st, _ := status.FromError(err)
	if st.Code() != codes.InvalidArgument {
		t.Fatalf("ожидался код InvalidArgument, получили %v", st.Code())
	}
}

func TestValidateToken_InvalidToken(t *testing.T) {
	ctx := context.Background()

	_, err := authClient.ValidateToken(ctx, &pb.ValidateRequest{
		AccessToken: "invalidtoken",
	})
	if err == nil {
		t.Fatal("ожидалась ошибка при неверном access token")
	}
	st, _ := status.FromError(err)
	if st.Code() != codes.InvalidArgument && st.Code() != codes.Unauthenticated {
		t.Fatalf("ожидался код InvalidArgument или Unauthenticated, получили %v", st.Code())
	}
}

func TestLogout_EmptyToken(t *testing.T) {
	ctx := context.Background()

	_, err := authClient.Logout(ctx, &pb.LogoutRequest{
		AccessToken: "",
	})
	if err == nil {
		t.Fatal("ожидалась ошибка при пустом access token")
	}
	st, _ := status.FromError(err)
	if st.Code() != codes.InvalidArgument {
		t.Fatalf("ожидался код InvalidArgument, получили %v", st.Code())
	}
}

func TestLogout_InvalidToken(t *testing.T) {
	ctx := context.Background()

	_, err := authClient.Logout(ctx, &pb.LogoutRequest{
		AccessToken: "invalidtoken",
	})
	if err == nil {
		t.Fatal("ожидалась ошибка при неверном access token")
	}
	st, _ := status.FromError(err)
	if st.Code() != codes.InvalidArgument && st.Code() != codes.Unauthenticated {
		t.Fatalf("ожидался код InvalidArgument или Unauthenticated, получили %v", st.Code())
	}
}

func TestCheckAdminStatus_EmptyUserId(t *testing.T) {
	ctx := context.Background()

	_, err := authClient.CheckAdminStatus(ctx, &pb.CheckAdminRequest{
		UserId: 0,
	})
	if err == nil {
		t.Fatal("ожидалась ошибка при пустом userId")
	}
	st, _ := status.FromError(err)
	if st.Code() != codes.InvalidArgument {
		t.Fatalf("ожидался код InvalidArgument, получили %v", st.Code())
	}
}
