# groupme.go
go sdk for groupme

## Summary
A simple sdk with no dependencies outside of the stdlib. 

## Usage
### Import
```go
import "github.com/nhomble/groupme.go/groupme"
```
### Send a message
```go
provider := groupme.TokenProviderFromToken("... your access token with groupme ....")
client, _ := groupme.NewClient(provider, nil)
client.Messages.Send(".. group id..", &groupme.SendMessageCommand{
		SourceGuid: "... guid ...",
		Text:       "Houston we have landed",
	})
```

## Examples
```sh
$ go run examples/sendMessage.go
```

You'll see here that the token is pulled from disk ```~/.groupme-go.prop``` and then we:
- create group
- create message
- list messages (and print our expectation)
- delete the group
