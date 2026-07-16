#!/usr/bin/env bash

commit_generated_code() {
    local model="$1"
    if [ -z "$model" ]; then
        read -r -p "What model has been used to generate the code?: " model
    fi
    git add .
    git commit -m "Add generated code" -m "Model: $model"
}
