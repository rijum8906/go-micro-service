set shell := ["zsh", "-c"]

default:
    @just --list

setup service:
    @just --justfile services/{{service}}/justfile setup
