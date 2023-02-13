package secrets_test

import (
	"context"
	"fmt"
	"os"

	"github.com/brexhq/substation/internal/secrets"
)

func ExampleGet_env() {
	// simulating a "secret" stored in an environment variable
	//nolint: tenv // not actually a test
	_ = os.Setenv("FOO", "bar")
	defer os.Unsetenv("FOO")

	// secrets stored in environment variables always begin with
	// "SECRETS_ENV" and end with the environment variable the
	// secret is in.
	secret, err := secrets.Get(context.TODO(), "SECRETS_ENV:FOO")
	if err != nil {
		// handle err
		panic(err)
	}

	// Output:
	// bar
	fmt.Println(secret)
}

func ExampleGet_aWS() {
	// secrets stored in AWS Secrets Manager always begin with
	// "SECRETS_AWS" and end with the name of the secret.
	secret, err := secrets.Get(context.TODO(), "SECRETS_AWS:FOO")
	if err != nil {
		// handle err
		panic(err)
	}

	fmt.Println(secret)
}
