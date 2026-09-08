module github.com/KARTIKrocks/gosms/examples/vonage-provider

go 1.27

require (
	github.com/KARTIKrocks/gosms v0.3.0
	github.com/KARTIKrocks/gosms/vonage v0.3.0
)

replace (
	github.com/KARTIKrocks/gosms => ../../
	github.com/KARTIKrocks/gosms/vonage => ../../vonage
)
