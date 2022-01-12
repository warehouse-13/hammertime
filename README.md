## Hammertime

A very basic, very untested, very unrefactored, and (likely) very unmaintained
CLI tool for interacting with [flintlock](https://github.com/weaveworks/flintlock) servers.

(Those of you who know your archaic gun mechanisms will understand the name and no doubt
find it hilarious.) (You are welcome.)

Why did I make this? Well we used to use another generic GRPC client, but we discovered
that, for reasons I do not yet know, it [didn't like some of the enum values we returned](https://github.com/weaveworks/flintlock/issues/313#issuecomment-991015159).
So here we are.

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

4 commands, very few configuration options. Everything has defaults so you don't need
to pass any flags at all if you don't want to. Each command simply spits out the response
as JSON so you can pipe to `jq` or whatever as you like.

```bash
# create 'mvm0' in 'ns0'
hammertime create

# get 'mvm0' in 'ns0'
hammertime get

# get just the state of 'mvm0' in 'ns0' *see below
hammertime get -s

# get all mvms in 'ns0'
hammertime list

# delete 'mvm0' from 'ns0'
hammertime delete
```

The name and namespace are configurable, as are the GRPC address and port.
There is the option to create with an SSH key.
You can also pass a full json configfile on the create if you want to override
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
