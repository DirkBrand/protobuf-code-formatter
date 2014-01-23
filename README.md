ProtoBuf-Code-Formatter
=======================

Code Formatter for Protocol Buffer 2.5.  Should be used as a stand-alone tool to format entire directories of code, or a plugin for protoc. 

To use the protofmt tool:

Install the tool, then run the following on the command-line:

`$ protofmt -r=true -proto_path='path' -exclude_path='list of paths' 'path of directory to format`

`-r` is a flag indicating whether to format the directory recursively or not.  
`-proto_path` is used to provide the location of all dependencies.  
`-exlude_path` is used to provide a colon separated list of paths of directories that should not be formatted.

The command will format and override all `.proto` files in the provided directory (not including the excluded directories).


For use in protoc:

Install the plugin, then have the location of the plugin binary in your PATH variable. Run the following on the command-line:

`$ protoc --pretty_out='location of output' 'location of unformatted .proto file' `

The command will format the input file and write it in the provided location.  If the location is the same as the original file, it will be overwritten.


Installation
============

To install the stand-alone tool, run the following command in the terminal:

` $ go get github.com/DirkBrand/protobuf-code-formatter/protofmt`

To install the plugin for protoc, run the following command in the terminal:

`$ go get github.com/DirkBrand/protobuf-code-formatter/protoc-gen-pretty`


Limitations
===========
1. Formatter cannot preserve order of structures

2. For comments, outer `extend' groups are logically grouped together, so inner comments are lost

3. Style of comments are not preserved (/* */ vs. //), so single-line comments are shown with `//` and multi-line comments with `/* */`.

4. When using the protoc plugin, any comments not directly adjacent to a line of code (dangling comments), are not preserved.  Comments must be directly above or below a line of code (without newlines).  Such comments are preserved when using the protofmt tool.


[![Build Status](https://drone.io/github.com/DirkBrand/protobuf-code-formatter/status.png)](https://drone.io/github.com/DirkBrand/protobuf-code-formatter/latest)
