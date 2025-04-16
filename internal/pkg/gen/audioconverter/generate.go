//go:generate protoc --go-grpc_out=require_unimplemented_servers=false:./../.. --go_out=./../.. protos/audio_convert.proto
package audioconverter
