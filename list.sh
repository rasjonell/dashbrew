#!/bin/env bash
ps -eo pcpu,comm --sort=-pcpu | head -n 6 | tail -n 5
