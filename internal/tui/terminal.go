   package tui

   import (
      "os"
      "golang.org/x/term"
   )

   func Size() (width, height int, err error) {
      return term.GetSize(int(os.Stdout.Fd()))
   }
