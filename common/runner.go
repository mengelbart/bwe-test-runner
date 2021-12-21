package common

type RunnerFactory func(date int64, testcase string, implementation string) Runner

type Runner interface {
	Run() error
}
