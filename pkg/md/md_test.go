package md

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestToJiraMD(t *testing.T) {
	jfm := `# H1
Some _Markdown_ text.

## H2
Foobar.

### H3
Fuga

> quote

- - - -

**strong text**
~~strikethrough text~~
[Example Domain](http://www.example.com/)
![](https://path.to/image.jpg)

* list1
* list2
* list3

Paragraph

1. number1
2. number2
3. number3

|a  |b  |c  |
|---|---|---|
|1  |2  |3  |
|4  |5  |6  |

{panel:title=My Title}
**Subtitle**

Some text with a title
{panel}

` + "```go" + `
package main

import "fmt"

func main() {
    fmt.Println("hello world")
}` + "```"

	expected := `h1. H1
Some _Markdown_ text.

h2. H2
Foobar.

h3. H3
Fuga

{quote}
quote

{quote}


----
*strong text*
-strikethrough text-
[Example Domain|http://www.example.com/]
!https://path.to/image.jpg!

* list1
* list2
* list3

Paragraph

# number1
# number2
# number3

||a||b||c||
|1|2|3|
|4|5|6|

{panel:title=My Title}
*Subtitle*

Some text with a title
{panel}

` + "```go" + `
package main

import "fmt"

func main() {
    fmt.Println("hello world")
}` + "```\n\n"

	assert.Equal(t, expected, ToJiraMD(jfm))
}

func TestToJiraMDDoesNotEscapePlainText(t *testing.T) {
	input := `Проверил (в скобках), из-за сбоя 2026-11-05, C++ и привет!`
	assert.Equal(t, input, ToJiraMD(input))
}

func TestToJiraMDKeepsAutolinksBare(t *testing.T) {
	assert.Equal(t, "https://example.com/browse/ABC-1", ToJiraMD("https://example.com/browse/ABC-1"))
}

func TestToJiraMDConvertsLabeledLinks(t *testing.T) {
	assert.Equal(t, "[Example|https://example.com/]\n\n", ToJiraMD("[Example](https://example.com/)"))
}
