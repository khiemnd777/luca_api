If `Linux` server

1. Check `fuser` PATH
```bash
which fuser

```
2. Open `sudoers` safely via:
```bash
sudo visudo
```

3. Add this code below at the end of file (if user is `administrator`) with `fuser` PATH above.
```bash
administrator ALL=(ALL) NOPASSWD: /usr/bin/fuser
```
