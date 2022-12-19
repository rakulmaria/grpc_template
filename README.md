# grpc_template

Very simple gRPC template setup for the exam. 

The service has a client and a server. The client can ask for the time and the server will respond with the time.Now() package. The service is defined in the proto.proto file

Very much based on [the walkthrough](https://github.com/theauk/grpcTimeRequestExample/blob/master/readme.md) with TA Thea.

## Important Notes

Every time you update the proto file, you have to run the protoc command (the long command in the proto file section).

    protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative grpc/proto.proto

Keep track of naming. For example, if the name of the service is changed to YearAsk then we e.g. also have to change the return statement in connectToServer to instead use proto.NewYearAskClient. This also goes for message and function changes.

Make sure you have imported "google.golang.org/grpc" in the server and ran go mod tidy.

If it doesn't work, try running

    go mod tidy

Ensure that you are using * and & properly. If you feel a bit shaky about this, you might want to consider using IntelliJ with the Go plugin, since it provides better intellisense.