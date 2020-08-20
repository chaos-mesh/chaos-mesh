## Notes:

* Keep in mind that the result will be cached by system!

  If you set read error with a probability < 1, once the program reads successfully, no error will be returned until the cache misses. If you override the attributes with a probability < 1, the first lookup may decide the attributes for a long time (until the cache misses).

  But if you set probability == 1, which means the result will be the same all the time during the mount, there will be no problem.

* Compile this binary with `-Z relro-level=full`, then it will load (mmap) all dependencies into memory at the beginning.