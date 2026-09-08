module github.com/KARTIKrocks/gosms/examples/msg91-provider

go 1.27

require (
	github.com/KARTIKrocks/gosms v0.2.1
	github.com/KARTIKrocks/gosms/msg91 v0.2.1
)

replace (
	github.com/KARTIKrocks/gosms => ../../
	github.com/KARTIKrocks/gosms/msg91 => ../../msg91
)
