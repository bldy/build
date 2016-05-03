load("//processor/tests/targetFromMacroWithDoubleLoadTHREE.BUILD", "LIB_COMPILER_FLAGS") 

harvey_library = cc_library(
	copts=LIB_COMPILER_FLAGS,
	includes=[
		"//sys/include",
		"//amd64/include",
	],
)
