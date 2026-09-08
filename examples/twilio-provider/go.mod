module github.com/KARTIKrocks/gosms/examples/twilio-provider

go 1.27

require (
	github.com/KARTIKrocks/gosms v0.3.0
	github.com/KARTIKrocks/gosms/twilio v0.3.0
)

replace (
	github.com/KARTIKrocks/gosms => ../../
	github.com/KARTIKrocks/gosms/twilio => ../../twilio
)
