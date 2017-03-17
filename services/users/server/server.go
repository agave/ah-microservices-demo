package server

import (
	"fmt"
	"net"

	log "github.com/Sirupsen/logrus"
	"github.com/agave/ah-microservices/services/users/db"
	"github.com/agave/ah-microservices/services/users/user"
	userGen "github.com/agave/ah-microservices/services/users/user/generated"
	"github.com/agave/ah-microservices/services/users/util"
	xContext "golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
)

// UserServer implements user service functionality through grpc
type userServer struct{}

// GetUser looks for the Id param in the database and returns user details if found
func (s *userServer) GetUser(ctx xContext.Context, id *userGen.Id) (*userGen.Profile, error) {
	log.WithFields(log.Fields{
		"GUID": id.GetGuid(), "data": id.String(),
	}).Info("Received request")

	if id.GetId() <= 0 {
		log.WithField("GUID", id.GetGuid()).Info("Not Found")
		return &userGen.Profile{}, grpc.Errorf(codes.NotFound, "Not Found")
	}

	log.WithField("GUID", id.GetGuid()).Info("Searching for user in db")

	u := user.Users{ID: id.GetId()}
	h, err := db.Engine.Get(&u)
	if err != nil {
		log.WithField("GUID", id.GetGuid()).Info("Internal DB Error")
		return &userGen.Profile{}, grpc.Errorf(codes.Internal, "%s", err.Error())
	}

	if h {
		log.WithField("GUID", id.GetGuid()).Info("Found")
		return &userGen.Profile{
			Id:      u.ID,
			Email:   u.Email,
			Balance: u.Balance,
		}, nil
	}

	log.WithField("GUID", id.GetGuid()).Info("Not Found")
	return &userGen.Profile{}, grpc.Errorf(codes.NotFound, "Not Found")
}

// CreateUser creates a user in the database based on the Create param
func (s *userServer) CreateUser(ctx xContext.Context, c *userGen.Create) (*userGen.Profile, error) {
	//TODO: email format/security validation
	if c.GetEmail() == "" {
		return &userGen.Profile{},
			grpc.Errorf(codes.InvalidArgument, "Email can't be empty")
	}

	if c.Balance < 0 {
		return &userGen.Profile{},
			grpc.Errorf(codes.InvalidArgument, "Balance can't be negative")
	}

	user := user.Users{Email: c.GetEmail()}
	exists, err := db.Engine.Get(&user)
	if err != nil {
		return &userGen.Profile{},
			grpc.Errorf(codes.Unknown, "Unable to create user")
	}

	if exists {
		return &userGen.Profile{},
			grpc.Errorf(codes.AlreadyExists, "User already exists")
	}

	user.Balance = c.GetBalance()

	// create user in db
	affectedRows, err := db.Engine.InsertOne(&user)
	if err == nil && affectedRows == 1 {
		return &userGen.Profile{
			Id:      user.ID,
			Email:   user.Email,
			Balance: user.Balance,
		}, nil
	}

	return &userGen.Profile{},
		grpc.Errorf(codes.Unknown, "Unable to create user")
}

func (s *userServer) VerifyUser(ctx xContext.Context,
	v *userGen.Verify) (*userGen.Verified, error) {
	GUIDID := log.Fields{"GUID": v.GetGuid(), "ID": v.Id.GetId()}
	log.WithFields(GUIDID).Debug("Verifying User")

	if v.Id.GetId() <= 0 {
		log.WithFields(GUIDID).Info("Not Found")
		return &userGen.Verified{CanUserFund: false},
			grpc.Errorf(codes.NotFound, "Not Found")
	}

	u := &user.Users{
		ID: v.Id.GetId(),
	}
	exists, err := db.Engine.Get(u)
	if err != nil {
		log.WithFields(log.Fields{
			"GUID": v.GetGuid(), "error": err.Error(),
		}).Info("Unable to get user")
		return &userGen.Verified{CanUserFund: false},
			grpc.Errorf(codes.Unknown, "%v", err)
	}

	if exists {
		return &userGen.Verified{CanUserFund: u.Balance >= v.GetAmount()}, nil
	}

	log.WithFields(GUIDID).Info("User not Found")
	return &userGen.Verified{CanUserFund: false},
		grpc.Errorf(codes.NotFound, "Not Found")
}

// StartServer configures and starts our Users grpc server
func StartServer() {
	// TODO: somehow test this
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", util.Config.Port))
	if err != nil {
		log.WithField("error", err).Fatalln("Failed to start GRPC Server")
	}
	grpcServer := grpc.NewServer()
	userGen.RegisterUserServer(grpcServer, &userServer{})
	grpcServer.Serve(lis)
}
