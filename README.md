# maws-cli

```console
$ maws -h
Usage of ./maws:
  -max-parallels int
        max parallels (default 10)
  -role-file string
        path of a role file
```

roles.txt
```txt
# comment
arn:aws:iam::123456789012:role/Foo
arn:aws:iam::012345678901:role/Bar
```

```console
$ maws -role-file roles.txt sts get-caller-identity
2021/05/21 17:15:48 [debug] assume role to arn:aws:iam::123456789012:role/Foo
2021/05/21 17:15:48 [debug] assume role to arn:aws:iam::012345678901:role/Bar
2021/05/21 17:15:48 [debug] ASIARSLAYHAF7LL4K4O2 [sts get-caller-identity]
2021/05/21 17:15:48 [debug] ASIA6MF6NY4IKWGJUVVP [sts get-caller-identity]
{
    "UserId": "AROAJPKWIXB7ME2EZEO3G:maws-1621584948",
    "Account": "314472643515",
    "Arn": "arn:aws:sts::123456789012:assumed-role/Foo/maws-1621584948"
}
{
    "UserId": "AROA6MF6NY4INNOYY5GST:maws-1621584948",
    "Account": "988241839888",
    "Arn": "arn:aws:sts::012345678901:assumed-role/Bar/maws-1621584948"
}
```
