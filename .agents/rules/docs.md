# Documentation

Prefix: DOC. User documentation lives in `docs/` and is published at https://psyb0t.github.io/gorpitx/.

- **DOC1:** `README.md` is the front door: what it is, install, one example, the module list, and links into the site. Reference material (fields, validation, examples per module) lives on the docs pages, once.
- **DOC2:** A change to a module's fields, defaults, validation or command line updates that module's docs page in the same change. A change to `Exec`, `Stop`, streaming or errors updates `docs/running-modules.md` or `docs/errors.md`.
- **DOC3:** Every page has front matter with a unique `title` and a `description` of 50 to 170 characters. Page-to-page links use the relative `.md` path. A link to a file outside `docs/` uses the full `https://github.com/psyb0t/gorpitx/blob/main/...` URL.
- **DOC4:** Run `make docs` and the site checker after editing `docs/`. No em dashes, "directory" not "folder".
