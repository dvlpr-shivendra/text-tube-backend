package helpers

import "strings"

func SanitizeMarkdown(input string) string {
    // Replace Non-Breaking Spaces (U+00A0) with standard spaces
    // Replace tabs with standard spaces
    // Ensure consistent newlines
    r := strings.NewReplacer(
        "\u00a0", " ", 
        "\u202f", " ", 
        "\r\n", "\n",
        "\n\n", " \n\n",
    )
    return r.Replace(input)
}