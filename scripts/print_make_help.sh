#!/bin/bash

set -eou pipefail

printf "Available targets:\n\n"
awk '/^[a-z]+/ {                                          # the line starts with a-z, so it might be a task definition
  if (match(prevLine, /^## (.*)/)) {                      # if the previous line started ##, we know the current line IS a task definition
    task_name = $1;
    gsub(":", "", task_name);                             # remove the trailing colon from the task name
    description = substr(prevLine, RSTART + 3, RLENGTH);  # get the description from the previous line
    is_subtask = gsub("/", "/", task_name);               # count the number of slashes in the task name
    if (is_subtask) { printf 1 } else { printf 0 };       # print 1 for subtasks, 0 for top-level tasks
    printf " %s %s\n", task_name, description;            # print the task name and description on the same line as the numeric identifier
  }
} { prevLine = $0; }' "$@" |                              # save the previous line so we can check it on the next iteration
sort -n |                                                 # partition tasks based on whether they are subtasks or not
cut -f 2- -d ' ' |                                        # remove the numeric identifier
awk '{
  printf "  \x1b[32;01m%-35s\x1b[0m ", $1;                # print the task name in green
  $1 = "";                                                # remove the task name from the line
  printf "%s\n", $0;                                      # print the remainder of the line
}'
