# Overview 

This utility reads a template file for SSM Parameter Store parameters, gets the values using the API, and outputs a new rendered file.

The system it runs on needs to have AWS CLI installed and configured
or have IAM access to the SSM Parameter Store
and access to the KMS key used to decrypt SecretStrings. The later would be the case if it is run on an EC2 instance where the instance profile has the correct permissions.
 
# Flags

```
-aws-region string
        the aws region where the parameters are stored defaults to AWS_DEFAULT_REGION  (default "us-east-1")
-output string
        name of the output file (default "output.txt")
-template string
        the template file to be used with format {{ <parameter name> }} (default "template.txt")
```

# Template File

The template file should use the format `{{ <parameter name> }}`
according to your naming convention. The spaces before and after the parameter name are required.

Example:

`DB_PASSWORD={{ /myapp/prod/DB_PASSWORD }}`

Will be replaced with:

`DB_PASSWORD=MySecretPassword`

# CI/CD

This utility is appropriate to be part of a CI/CD system.
It exits with status 1 and prints FAILURE which is recognized by
most CI/CD systems and will stop the build.

