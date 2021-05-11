package main

import (
	"os"

	refactor "gitee.com/azhai/xorm-refactor"
	"gitee.com/azhai/xorm-refactor/cmd"
	"github.com/urfave/cli/v2"
)

func main() {
	app := &cli.App{
		Version: cmd.VERSION,
		Usage:   "从数据库导出对应的Model代码",
		Action:  ReverseAction,
	}
	app.Flags = []cli.Flag{
		&cli.BoolFlag{
			Name:    "verbose",
			Aliases: []string{"vv"},
			Usage:   "输出详细信息",
		},
		&cli.StringFlag{
			Name:    "file",
			Aliases: []string{"c", "f"},
			Usage:   "配置文件路径",
			Value:   "settings.yml",
		},
		&cli.StringFlag{
			Name:    "namespace",
			Aliases: []string{"ns"},
			Usage:   "项目NameSpace",
			Value:   "xorm-refactor",
		},
	}
	err := app.Run(os.Args)
	if err != nil {
		panic(err)
	}
}

func ReverseAction(ctx *cli.Context) error {
	configFile := ctx.String("file")
	nameSpace := ctx.String("namespace")
	settings, err := cmd.Prepare(configFile, nameSpace)
	if err != nil {
		return err
	}
	names := ctx.Args().Slice()
	verbose := cmd.Verbose() || ctx.Bool("verbose")
	err = refactor.ExecReverseSettings(settings, verbose, names...)
	return err
}
