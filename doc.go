/*

Package presenter implements parsing and rendering of present files,
which can be slide presentations as in github.com/coding-socks/presenter/cmd/present.

File Format

Present files begin with a header giving the title of the document
and other metadata, which looks like:

	# Title of document
	Subtitle of document
	15:04 2 Jan 2006
	Tags: foo, bar, baz
	Summary: This is a great document you want to read.

The "#" indicates the title. Between the title and other metadata must be
no empty lines.

The date line may be written without a time:
	2 Jan 2006
In this case, the time will be interpreted as 11am UTC on that date.

The tags line is a comma-separated list of tags that may be used to categorize
the document.

The summary line gives a short summary used in blog feeds.

Only the title is required;
the subtitle, date, tags, and summary lines are optional.

After the header come zero or more author blocks, like this:

	Author Name
	Job title, Company
	<joe@example.com>
	<https://url/>
	<@twitter_name>

The first line of the author block is conventionally the author name.
Otherwise, the author section may contain a mixture of text, twitter names, and links.
For slide presentations, only the plain text lines will be displayed on the
first slide.

If multiple author blocks are listed, each new block must be preceded
by its own blank line.

After the author blocks come the presentation slides or article sections,
which can in turn have subsections.
Each slide or section begins with "##" or "###" header line. As of now "####" or more are
handled as part of the slide content.

Markdown Syntax

Markdown typically means the generic name for a family of similar markup languages.
The specific variant used in present is CommonMark with some extensions to support
strikethrough, task list, and twitter link components.
See https://commonmark.org/help/tutorial/ for a quick tutorial.

Example:

	# Title of document
	Subtitle of document
	15:04 2 Jan 2006
	Tags: foo, bar, baz
	Summary: This is a great document you want to read.

	My Name
	<me@example.com>

	## Title of 2nd level Slide (must begin with ##)

	Some Text

	### Title of 3rd level Slide (must begin with ###)

	More text

	## Single slide (must have empty body)

	### Example formatting

	Formatting:

	The following text formatted as _italic_, **bold**,
	`code`, ~~deleted~~.

	Links can have different formats as well.
	[This link is labeled with text](https://example.com/),
	On the other hand, the following is labeled with the link itself
	<https://example.com/>. The next one is an email address <me@example.com>.
	And finally, a twitter alias <@rob_pike>.

*/
package presenter
