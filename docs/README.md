# gh-contrib

- [Features](#features)
- [Usage](#usage)
- [Installation](#installation)

This is a GitHub CLI extension that allows you to retrieve and display a summary of your contributions to GitHub within a specified date range.

## Features

- Retrieve contribution data for a specific date range.
- Display contributions in a tabular format with date, contribution level, and contribution count.
- Easy-to-use command-line interface.

![demo.gif](./demo.gif)

## Usage

You can use this extension to retrieve your contribution summary by specifying a date range using the from and to flags. For example, to retrieve contributions from September 1, 2023, to September 30, 2023:

```shell
gh contrib --from 2023-09-01 --to 2023-09-30
```

- `--from`: The start date for the contribution summary in the format YYYY-MM-DD. Defaults to 5 days ago if not specified.
- `--to`: The end date for the contribution summary in the format YYYY-MM-DD. Defaults to today if not specified.

## Installation

To install this GitHub CLI extension, use the following command.

```shell
gh extension install kmtym1998/gh-contrib
```
