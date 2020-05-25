@ECHO OFF

REM USAGE:
REM     make      Rebuild the project

IF "%1" == "" GOTO all
IF "%1" == "all" GOTO all
IF "%1" == "build" GOTO build

:all
    del refactor.exe
:build
    go.exe build -ldflags="-s -w" -o refactor.exe ./cmd/reverse/
