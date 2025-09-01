# MIT License

Copyright (c) 2024 STIG Viewer X Contributors

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.

---

## Third-Party Licenses

This project includes or depends on software with the following licenses:

### Go Dependencies
- **github.com/google/uuid**: BSD-3-Clause
- **github.com/spf13/cobra**: Apache-2.0
- **github.com/wailsapp/wails/v2**: MIT
- **modernc.org/sqlite**: BSD-3-Clause
- **golang.org/x/crypto**: BSD-3-Clause
- **golang.org/x/net**: BSD-3-Clause
- **golang.org/x/sys**: BSD-3-Clause
- **golang.org/x/text**: BSD-3-Clause

### Frontend Dependencies
- **React**: MIT
- **TypeScript**: Apache-2.0
- **Vite**: MIT

For full license texts of third-party dependencies, see the respective project repositories or run:
```bash
go mod download && go list -m -json all
```

## License Compliance

This project aims to maintain license compatibility and transparency. All dependencies are chosen to ensure compatibility with the MIT license and DoD requirements for open-source software.

If you have questions about licensing or need clarification about third-party dependencies, please open an issue on GitHub.