module github.com/KARTIKrocks/gosms/examples/twilio-provider

go 1.24

require (
	github.com/KARTIKrocks/gosms v0.2.0
	github.com/KARTIKrocks/gosms/twilio v0.2.0
)

replace (
	github.com/KARTIKrocks/gosms => ../../
	github.com/KARTIKrocks/gosms/twilio => ../../twilio
)
