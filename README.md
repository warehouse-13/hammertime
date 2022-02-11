## Hammertime

A very basic CLI tool for interacting with [flintlock](https://github.com/weaveworks/flintlock) servers.

(Those of you who know your archaic gun mechanisms will understand the name and no doubt
find it hilarious.) (You are welcome.)

Why did I make this? Well we used to use another generic GRPC client, but we discovered
that, for reasons I do not yet know, it [didn't like some of the enum values we returned](https://github.com/weaveworks/flintlock/issues/313#issuecomment-991015159).
So here we are.

I have kept it around because it makes working with flintlock very straightforward.

### Versioning

Latest of hammertime is always aligned with latest of flintlock.
Check the release notes for potential breakages.

TODO compatibility table

### Installation

1. Build from source:
   ```bash
   git clone <this or your fork>
   cd hammertime
   make build
   ```

2. Get a [released binary](https://github.com/Callisto13/hammertime/releases) (linux only)

3. Install with go: `go install github.com/Callisto13/hammertime/releases@latest`


Alias to `ht` if you like.

### Usage

4 commands, very few configuration options. Each command simply spits out the response
as JSON so you can pipe to `jq` or whatever as you like.

```bash
# create 'mvm0' in 'ns0' (take note of the UID after creation)
hammertime create

# get
hammertime get -i <UUID>

# get just the state of 'mvm0' in 'ns0' *see below
hammertime get -i <UUID> -s

# get all mvms across all namespaces
hammertime list

# delete
hammertime delete -i <UID>

# delete all mvms everywhere
hammertime clear
```

The name and namespace are configurable, as are the GRPC address and port.
There is the option to create with an SSH key.
You can also pass a full json configfile to `create`, `get` and `delete` if you want to override
everything (see [example.json](example.json)).

Run `hammertime --help` for details.

\* Why have a specific flag for getting the state? Why not just let users do whatever
with `jq` or equivalent? Well, when the state is `PENDING` it is enum value `0`
in our proto, which ends up equating to a `null` value and so that is what ends
up being populated (or doesn't really, but get it).
When you call it explicitly on the received object, the client will understand
that it is not actually `null` and translate the `0` properly.
Furthermore, even when
the value is set to some non-zero state, then all that will be in the printed result
is the enum number, which is not very user friendly. So this is why we have the
flag. I could totally do a conversion
func which would solve both these issues, but I cnba right now.


### Development

#### Testing

Test can be run with `make test`.
For a list of all make commands, run `make help`.
