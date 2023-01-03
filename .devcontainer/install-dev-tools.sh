# download go packages
go mod download
# compile protobuf
sh build/scripts/proto/compile.sh
# compile configurations
sh build/scripts/config/compile.sh
