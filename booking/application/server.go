package application

import (
	"booking/domain"
	pb "booking/pb"
	"context"

	"github.com/golang/protobuf/ptypes"
	"github.com/golang/protobuf/ptypes/empty"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type grpcServer struct {
	service Service
	pb.UnimplementedBookingServiceServer
}

// NewServer provides instance of grpc server.
func NewServer(s Service) pb.BookingServiceServer {
	return &grpcServer{s, pb.UnimplementedBookingServiceServer{}}
}

func (s *grpcServer) BookNewCargo(_ context.Context, req *pb.BookNewCargoRequest) (*pb.BookNewCargoResponse, error) {
	origin := domain.UNLocode(req.GetOrigin())
	destination := domain.UNLocode(req.GetDestination())
	deadline, err := ptypes.Timestamp(req.GetDeadline())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	trackingID, err := s.service.BookNewCargo(origin, destination, deadline)
	if err != nil {
		return nil, status.Error(codes.Unknown, err.Error())
	}
	return &pb.BookNewCargoResponse{TrackingId: string(trackingID)}, nil
}

func (s *grpcServer) RequestPossibleRoutesForCargo(req *pb.RequestPossibleRoutesForCargoRequest, stream pb.BookingService_RequestPossibleRoutesForCargoServer) error {
	id := domain.TrackingID(req.GetTrackingId())
	itineraries := s.service.RequestPossibleRoutesForCargo(id)

	for _, itin := range itineraries {
		pbItin, err := encodeItinerary(&itin)
		if err != nil {
			return status.Error(codes.Unknown, err.Error())
		}
		stream.Send(pbItin)
	}

	return nil
}

func (s *grpcServer) AssignCargoToRoute(_ context.Context, req *pb.AssignCargoToRouteRequest) (*empty.Empty, error) {
	id := domain.TrackingID(req.GetTrackingId())
	itinerary, err := decodeItinerary(req.GetItinerary())
	if err != nil {
		return nil, status.Error(codes.Unknown, err.Error())
	}
	err = s.service.AssignCargoToRoute(id, *itinerary)
	if err != nil {
		return nil, status.Error(codes.Unknown, err.Error())
	}
	return &empty.Empty{}, nil
}

func (s *grpcServer) ChangeDestination(_ context.Context, req *pb.ChangeDestinationRequest) (*empty.Empty, error) {
	id := domain.TrackingID(req.GetTrackingId())
	dest := domain.UNLocode(req.GetDestination())
	if err := s.service.ChangeDestination(id, dest); err != nil {
		return nil, status.Error(codes.Unknown, err.Error())
	}
	return &empty.Empty{}, nil
}

func (s *grpcServer) Locations(_ context.Context, _ *empty.Empty) (*pb.LocationsResponse, error) {
	locations := s.service.Locations()
	pbLocations := []*pb.LocationsResponse_Location{}
	for _, loc := range locations {
		pbLocations = append(pbLocations, &pb.LocationsResponse_Location{
			UnLocode: loc.UNLocode,
			Name:     loc.Name,
		})
	}
	return &pb.LocationsResponse{Locations: pbLocations}, nil
}

func (s *grpcServer) LoadCargo(_ context.Context, req *pb.LoadCargoRequest) (*pb.LoadCargoResponse, error) {
	id := domain.TrackingID(req.GetTrackingId())
	cargo, err := s.service.LoadCargo(id)
	if err != nil {
		return nil, status.Error(codes.Unknown, err.Error())
	}

	pbCargo, err := encodeCargo(&cargo)
	if err != nil {
		return nil, status.Error(codes.Unknown, err.Error())
	}

	return &pb.LoadCargoResponse{Cargo: pbCargo}, nil
}

func (s *grpcServer) Cargos(_ context.Context, _ *empty.Empty) (*pb.CargosResponse, error) {
	cargos := s.service.Cargos()
	pbCargos := make([]*pb.Cargo, 0, len(cargos))
	for _, cargo := range cargos {
		pbCargo, err := encodeCargo(&cargo)
		if err != nil {
			return nil, status.Error(codes.Unknown, err.Error())
		}
		pbCargos = append(pbCargos, pbCargo)
	}
	return &pb.CargosResponse{Cargos: pbCargos}, nil
}
