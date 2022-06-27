## Integration tests

The integration tests are very high level.
They only test the core user stories.
For more granular tests for every config and flag variation, add a unit test.

### How to run

```
make int
```

To run against a real flintlock instance, start your flintlockd server then run:

```
TEST_SERVER="192.168.0.31:9091" make int
```
