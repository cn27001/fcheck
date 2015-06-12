### fcheck

Is a simple file checking utility that is meant to detect any changes that happen to a filesystem.

The idea is that fcheck is used to generate a db, that will store information about a filesystem, and fcheck is then re-run later
to detect any tempering with files. It does this by storing things such as permissions, size, last modified date, 
and perhaps most importantly sha512sum checksum of the file contents (for regular files larger than 0).

To show usage:

`./fcheck -h` 

To generate the db:

`./fcheck -path=/ -gendb -exclude_from=excludes.txt`

To check integrity of a filesystem against previously generated db:

`./fcheck -path=/ -exclude_from=excludes.txt`

The `-excludes_from` can be  omitted as it defaults to excludes.txt.


Sample excludes.txt

```
/dev
/mnt
/proc
/sys
/tmp
```

To display an entry in fcheck's db (e.g. when trying to figure out how /bin/ps was tempered with)

`./fcheck -path=/bin/ps -show`

Personally after generating the db I move/copy both the fcheck binary and the fcheck.db and fcheck.db.index onto a removable device.
For added peace of mind I sign them with `gpg -b`. And then later mount that device read-only to detect any changes to my filesystem.
