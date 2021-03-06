package injection

import (
	"fmt"
	"github.com/MaximZayats/godi/codegen"
	"github.com/MaximZayats/godi/di"
	"log"
	"os"
)

var (
	config            = codegen.DefaultConfig
	injectedFunctions = make([]injectedFunction, 0)
)

type injectedFunction struct {
	inFunc        any
	outFunc       any
	wasFound      bool
	wasSuccessful bool
	details       string
}

func Configure(c codegen.Config) { config = c }

func Inject[OutFunc any, InFunc any](
	f InFunc, container ...*di.Container,
) OutFunc {
	decorator, ok := config.GetterFunction(f)
	if !ok {
		injectedFunctions = append(injectedFunctions, injectedFunction{
			inFunc:   *new(InFunc),
			outFunc:  *new(OutFunc),
			wasFound: false,
			details:  "Not found",
		})
		return *new(OutFunc)
	}

	typedDecorator, ok := decorator.(func(InFunc, *di.Container) any)
	if !ok {
		injectedFunctions = append(injectedFunctions, injectedFunction{
			inFunc:        *new(InFunc),
			outFunc:       *new(OutFunc),
			wasFound:      true,
			wasSuccessful: false,
			details:       "`typedDecorator` type missmatch",
		})
		return *new(OutFunc)
	}

	c := di.GetContainer(container...)
	decoratedFunction, ok := typedDecorator(f, c).(OutFunc)
	if !ok {
		injectedFunctions = append(injectedFunctions, injectedFunction{
			inFunc:        *new(InFunc),
			outFunc:       *new(OutFunc),
			wasFound:      true,
			wasSuccessful: false,
			details:       "`decoratedFunction` type missmatch",
		})
		return *new(OutFunc)
	}

	injectedFunctions = append(injectedFunctions, injectedFunction{
		inFunc:        *new(InFunc),
		outFunc:       *new(OutFunc),
		wasFound:      true,
		wasSuccessful: true,
	})

	return decoratedFunction

}

func VerifyInjections() bool {
	numberOfUnInjectedFunctions := 0
	signatures := make([]codegen.Signature, 0)

	for _, value := range injectedFunctions {
		signatures = append(signatures, codegen.NewSignature(value.inFunc, value.outFunc))
		if !value.wasFound || !value.wasSuccessful {
			numberOfUnInjectedFunctions += 1
		}
	}

	if numberOfUnInjectedFunctions != 0 {
		fmt.Printf("Found %d unupdated functions\n", numberOfUnInjectedFunctions)
		fmt.Printf("Regenerating %d decorators...\n", len(signatures))

		err := codegen.Generate(config, signatures...)
		if err != nil {
			fmt.Println("Error while generating...")
			log.Fatal(err)
		}

		fmt.Printf("%d decorators was successfully regenerated!\n", len(signatures))

		return false
	}

	return true
}

func MustVerifyInjections() {
	ok := VerifyInjections()
	if !ok {
		fmt.Println(
			"The injection functions have been changed.\n" +
				"A restart is required.",
		)
		os.Exit(0)
	}
}
