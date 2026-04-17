// Package backends contains the collection of built-in backends that are supported.
//
// To register the built-in backends, you **must** import the one of the sub-packages in this package:
//
// For example:
//
//		 import (
//	  	 _ "go.innotegrity.dev/mod/uripath/backends/aws"
//	  	 _ "go.innotegrity.dev/mod/uripath/backends/generic"
//	  	 _ "go.innotegrity.dev/mod/uripath/backends/google"
//	  	 _ "go.innotegrity.dev/mod/uripath/backends/hashicorp"
//	  )
package backends
