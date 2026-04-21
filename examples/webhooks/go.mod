module github.com/KARTIKrocks/gosms/examples/webhooks

go 1.24

require (
	github.com/KARTIKrocks/gosms v0.1.0
	github.com/KARTIKrocks/gosms/msg91 v0.1.0
	github.com/KARTIKrocks/gosms/twilio v0.1.0
	github.com/KARTIKrocks/gosms/vonage v0.1.0
)

replace (
	github.com/KARTIKrocks/gosms => ../../
	github.com/KARTIKrocks/gosms/msg91 => ../../msg91
	github.com/KARTIKrocks/gosms/twilio => ../../twilio
	github.com/KARTIKrocks/gosms/vonage => ../../vonage
)
