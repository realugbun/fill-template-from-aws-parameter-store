package main

/*

This application is designed to be part of a CI/CD system.
It exists with status 1 and prints FAILURE which is recognized by
most CI/CD systems and will stop the build.

The system it runs on needs to have AWS CLI installed and configured
or have IAM access to the SSM Parameter Store
and access to the KMS key used to decrypt SecretStrings.

The template file should use the format {{ <parameter name> }}
according to your naming convention.

Example:

DB_PASSWORD={{ /myapp/prod/DB_PASSWORD }}

Will be replaced with:

DB_PASSWORD=MySecretPassword

*/

import (
	"flag"
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ssm"
)

func main() {

	// Setup flags
	t := flag.String("template", "template.txt", "the template file to be used with format {{ <parameter name> }}")
	o := flag.String("output", "output.txt", "name of the output file")
	r := flag.String("aws-region", os.Getenv("AWS_DEFAULT_REGION"), "the aws region where the parameters are stored defaults to AWS_DEFAULT_REGION ")

	flag.Parse()

	// Map holding the key and value of each parameter
	parameters := map[string]string{}

	// Open the template file
	template, err := os.ReadFile(*t)
	if err != nil {
		fmt.Println("FAILURE: ", err)
		os.Exit(1)
	}

	// Find all the template items
	re := regexp.MustCompile(`{{ (.*) }}`)
	matched := re.FindAllString(string(template), -1)
	for _, v := range matched {
		// Remove the template prefix and suffix to get just the parameter name.
		v = strings.ReplaceAll(v, `{{ `, "")
		v = strings.ReplaceAll(v, ` }}`, "")
		// Add the parameter name as a key in the map with an empty value
		parameters[v] = ""
	}

	// Create the input parameters we will send to SSM
	getParametersInput := &ssm.GetParametersInput{
		WithDecryption: aws.Bool(true),
	}

	// Add the name of the parameters we want to the request
	for k := range parameters {
		key := k
		getParametersInput.Names = append(getParametersInput.Names, &key)
	}

	// Create a session with the region we want from the region flag
	sess, err := session.NewSessionWithOptions(session.Options{
		Config:            aws.Config{Region: aws.String(*r)},
		SharedConfigState: session.SharedConfigEnable,
	})
	if err != nil {
		fmt.Println("FAILURE: ", err)
		os.Exit(1)
	}

	// Call SSM
	ssmsvc := ssm.New(sess, aws.NewConfig())
	param, err := ssmsvc.GetParameters(getParametersInput)
	if err != nil {
		fmt.Println("FAILURE: ", err)
		os.Exit(1)
	}

	// If any parameters are not found in the store,
	// the API does NOT return the ones that were found.
	// It only returns a list of the invalid parameters.
	// Check for this and exit.
	if len(param.InvalidParameters) > 0 {
		fmt.Println("FAILURE: Requested parameters do not exist on SSM.")
		for _, v := range param.InvalidParameters {
			fmt.Printf(" - %v\n", *v)
		}
		os.Exit(1)
	}

	// Put the values we got from the API into our map
	for _, p := range param.Parameters {
		template = []byte(strings.ReplaceAll(string(template), "{{ "+*p.Name+" }}", *p.Value))
	}

	// Write out the new file
	os.WriteFile(*o, template, 0644)
}
