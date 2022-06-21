module idefix-go

go 1.18

require (
	github.com/eclipse/paho.mqtt.golang v1.4.1
	github.com/stretchr/testify v1.7.4
	github.com/vmihailenco/msgpack/v5 v5.3.5
	gitlab.com/garagemakers/idefix-go/minips v0.0.0
	gitlab.com/garagemakers/idefix/core/cert v0.0.0-20220610083502-8dd9faf3adf2
	gitlab.com/garagemakers/idefix/core/idefix v0.0.9
)

require (
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/gorilla/websocket v1.5.0 // indirect
	github.com/jaracil/ei v0.0.0-20170808175009-4f519a480ebd // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/vmihailenco/tagparser/v2 v2.0.0 // indirect
	golang.org/x/net v0.0.0-20220607020251-c690dde0001d // indirect
	golang.org/x/sync v0.0.0-20220601150217-0de741cfad7f // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace gitlab.com/garagemakers/idefix-go/minips => ./minips
