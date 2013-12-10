include $(GOROOT)/src/Make.$(GOARCH)

	TARG=path/to/example
	GOFILES=\
		test.pb.go\
		other.go

	include $(GOROOT)/src/Make.pkg
	include $(GOROOT)/src/pkg/code.google.com/p/goprotobuf/Make.protobuf
