package prompts

// CriterionPrompt contains the evaluation prompt for a specific criterion.
type CriterionPrompt struct {
	ID          string
	Name        string
	Description string
	Prompt      string
	Examples    []Example
}

// Example provides few-shot examples for the LLM.
type Example struct {
	Input  string
	Output string
}

// CriterionPrompts maps criterion IDs to their evaluation prompts.
var CriterionPrompts = map[string]CriterionPrompt{
	"1.2.1": {
		ID:          "1.2.1",
		Name:        "Audio-only and Video-only (Prerecorded)",
		Description: "Prerecorded audio-only and video-only media have alternatives.",
		Prompt: `Evaluate whether this audio-only or video-only media has an adequate alternative.

Media element: {{.Element.OuterHTML}}
Media type: {{.MediaType}}
Transcript/alternative found: {{.TranscriptContent}}
Surrounding context: {{.SurroundingContext}}

For AUDIO-ONLY content (podcasts, audio recordings):
1. Is there a text transcript nearby or linked?
2. Does the transcript capture all spoken content?
3. Are speaker identifications included?
4. Are relevant sounds described (laughter, applause, music)?

For VIDEO-ONLY content (silent videos, animations):
1. Is there a text description or audio description?
2. Does the alternative convey the visual information?
3. Are actions, settings, and visual changes described?

If transcript/alternative is provided, evaluate if it is EQUIVALENT to the media content.
Use STT transcription (if available) to verify transcript accuracy.`,
	},

	"1.2.2": {
		ID:          "1.2.2",
		Name:        "Captions (Prerecorded)",
		Description: "Captions are provided for all prerecorded audio content in synchronized media.",
		Prompt: `Evaluate whether the captions for this video are adequate.

Video element: {{.Element.OuterHTML}}
Caption track found: {{.CaptionTrackPresent}}
Caption sample: {{.CaptionSample}}
STT transcription sample: {{.STTTranscription}}

Evaluate caption quality:
1. ACCURACY: Do captions match the spoken audio? Compare to STT output.
2. SYNCHRONIZATION: Are captions timed correctly with speech?
3. COMPLETENESS: Are all speakers captioned? Is relevant audio (music, sounds) noted?
4. SPEAKER IDENTIFICATION: Are different speakers labeled?
5. READABILITY: Is caption timing appropriate for reading?

Note: axe-core checks for caption PRESENCE. This evaluation checks caption QUALITY.
If no captions are present, this is an automatic failure.`,
	},

	"1.2.3": {
		ID:          "1.2.3",
		Name:        "Audio Description or Media Alternative (Prerecorded)",
		Description: "An alternative for time-based media or audio description is provided.",
		Prompt: `Evaluate whether this video has adequate audio description or media alternative.

Video element: {{.Element.OuterHTML}}
Audio description track: {{.AudioDescriptionTrack}}
Media alternative (transcript): {{.MediaAlternative}}
{{if .Screenshot}}Video frame sample: [See screenshot]{{end}}
{{if .VideoDescription}}AI video description: {{.VideoDescription}}{{end}}

For AUDIO DESCRIPTION (separate audio track):
1. Are important visual elements described during natural pauses?
2. Are actions, expressions, and scene changes conveyed?
3. Is the description synchronized appropriately?

For MEDIA ALTERNATIVE (full text transcript):
1. Does it include all dialogue AND visual information?
2. Are visual actions and settings described?
3. Would someone understand the full content from text alone?

Compare the provided alternative against the AI-generated video description.
Flag if significant visual information is missing from the alternative.`,
	},

	"1.2.5": {
		ID:          "1.2.5",
		Name:        "Audio Description (Prerecorded)",
		Description: "Audio description is provided for all prerecorded video content.",
		Prompt: `Evaluate whether this video has adequate audio description.

Video element: {{.Element.OuterHTML}}
Audio description track: {{.AudioDescriptionTrack}}
{{if .Screenshot}}Video frame samples: [See screenshots]{{end}}
{{if .VideoDescription}}AI video description: {{.VideoDescription}}{{end}}

Audio description requirements (stricter than 1.2.3):
1. Is there a dedicated audio description track (not just captions)?
2. Does it describe visual information not in dialogue?
3. Are key visual elements described:
   - Character actions and expressions
   - Scene changes and settings
   - On-screen text and graphics
   - Visual plot points

Compare the audio description against AI-generated video analysis.
Mark as FAIL if no audio description track exists.
Mark as NEEDS_REVIEW if description seems incomplete.`,
	},

	"1.1.1": {
		ID:          "1.1.1",
		Name:        "Non-text Content",
		Description: "All non-text content has a text alternative that serves the equivalent purpose.",
		Prompt: `Evaluate whether this non-text content has an adequate text alternative.

Element: {{.Element.OuterHTML}}
Alt text: {{.Element.Attributes.alt}}
Accessible name: {{.Element.AccessibleName}}
Surrounding context: {{.SurroundingContext}}

Consider:
1. Does the alt text convey the same information/purpose as the image?
2. Is it descriptive enough for someone who cannot see the image?
3. For decorative images, is it properly marked as such (empty alt="")?
4. For complex images (charts, diagrams), is there extended description?

If you can see a screenshot, use it to verify the alt text is accurate.`,
	},

	"1.3.2": {
		ID:          "1.3.2",
		Name:        "Meaningful Sequence",
		Description: "When the sequence in which content is presented affects its meaning, a correct reading sequence can be programmatically determined.",
		Prompt: `Evaluate whether the content reading sequence is meaningful.

DOM order (first 20 elements): {{.HTML}}
Visual layout (if screenshot provided): {{if .Screenshot}}[See screenshot]{{else}}Not available{{end}}

Consider:
1. Does the DOM order match the intended reading order?
2. Would a screen reader user understand the content in the order it's presented?
3. Are there any elements positioned visually differently than their DOM order?
4. Could CSS positioning cause confusion about reading order?`,
	},

	"1.3.3": {
		ID:          "1.3.3",
		Name:        "Sensory Characteristics",
		Description: "Instructions provided for understanding and operating content do not rely solely on sensory characteristics.",
		Prompt: `Evaluate whether instructions rely solely on sensory characteristics.

Text content: {{.HTML}}

Look for instructions that reference ONLY:
- Shape: "Click the round button", "Find the square icon"
- Color: "Click the red button", "Required fields are in red"
- Size: "Click the large button"
- Location: "Use the menu on the left", "Click the button below"
- Sound: "Click when you hear the beep"

Instructions are acceptable if they ALSO include non-sensory cues like:
- Text labels: "Click the Submit button (it's green)"
- Names: "Click the Search button on the left"`,
	},

	"1.4.1": {
		ID:          "1.4.1",
		Name:        "Use of Color",
		Description: "Color is not used as the only visual means of conveying information, indicating an action, prompting a response, or distinguishing a visual element.",
		Prompt: `Evaluate whether color is the only means of conveying information.

Page HTML: {{.HTML}}
{{if .Screenshot}}Visual reference: [See screenshot]{{end}}

Look for:
1. Links that are only distinguished by color (no underline, bold, or other indicator)
2. Required fields marked only with color (red asterisk without text explanation)
3. Error states shown only with color (red border without text/icon)
4. Status indicators using only color (green=success, red=error without labels)
5. Charts/graphs where data series are only distinguished by color
6. Form validation that only uses color to indicate success/failure

Acceptable patterns:
- Color PLUS text labels
- Color PLUS icons/shapes
- Color PLUS patterns/textures
- Links with color AND underline (on hover counts)`,
	},

	"1.4.5": {
		ID:          "1.4.5",
		Name:        "Images of Text",
		Description: "If the technologies being used can achieve the visual presentation, text is used to convey information rather than images of text.",
		Prompt: `Evaluate whether images of text are used unnecessarily.

Page HTML: {{.HTML}}
{{if .Screenshot}}Visual reference: [See screenshot]{{end}}

Look for:
1. Images containing text that could be rendered as HTML/CSS
2. Logos and branding (these are EXEMPT - acceptable as images)
3. Text in buttons/headers that's actually an image
4. Screenshots of text content
5. Infographics where text could be separate

Images of text are acceptable when:
- The text is part of a logo or brand name
- The specific presentation is essential (artistic typography)
- The image is customizable by the user

Images of text are NOT acceptable when:
- The same visual effect could be achieved with CSS
- It's regular body text or navigation
- It prevents text resizing or causes accessibility issues`,
	},

	"2.4.4": {
		ID:          "2.4.4",
		Name:        "Link Purpose (In Context)",
		Description: "The purpose of each link can be determined from the link text alone or together with its programmatically determinable context.",
		Prompt: `Evaluate whether the link purpose is clear from context.

Link element: {{.Element.OuterHTML}}
Link text: {{.Element.TextContent}}
Accessible name: {{.Element.AccessibleName}}
Surrounding context: {{.SurroundingContext}}

Consider:
1. Can you understand where this link goes from the text alone?
2. If not, does the surrounding paragraph/list/table provide context?
3. Would "Read more" or "Click here" be confusing without context?
4. Is there an aria-label or aria-describedby that helps?`,
	},

	"2.4.5": {
		ID:          "2.4.5",
		Name:        "Multiple Ways",
		Description: "More than one way is available to locate a Web page within a set of Web pages except where the Web Page is the result of, or a step in, a process.",
		Prompt: `Evaluate whether multiple ways exist to locate pages within the site.

Page HTML: {{.HTML}}

Look for at least TWO of these navigation methods:
1. Site navigation menu (header/footer nav)
2. Site search functionality
3. Site map page
4. Table of contents
5. List of related pages
6. Breadcrumb navigation
7. A-Z index

Exemptions (single navigation method acceptable):
- Result pages from a search
- Steps in a multi-step process (checkout, wizard)
- Confirmation pages after form submission

Note: This criterion applies to pages within a "set" - standalone single-page sites may be exempt.`,
	},

	"2.4.6": {
		ID:          "2.4.6",
		Name:        "Headings and Labels",
		Description: "Headings and labels describe topic or purpose.",
		Prompt: `Evaluate whether headings and labels describe their topic or purpose.

Page HTML: {{.HTML}}

Examine all headings (h1-h6) and form labels:

For headings, check:
1. Do they describe the content that follows?
2. Are they specific enough to be meaningful?
3. Would a user scanning headings understand the page structure?
4. Avoid: "Introduction", "More Info", "Details" without context

For labels, check:
1. Do they clearly identify the input purpose?
2. Are they visible (not just placeholder text)?
3. Would a user know what to enter?
4. Avoid: unlabeled inputs, icons-only labels

Note: Empty headings are a separate issue (automated check). This criterion focuses on descriptive quality.`,
	},

	"2.5.7": {
		ID:          "2.5.7",
		Name:        "Dragging Movements",
		Description: "All functionality that uses a dragging movement for operation can be achieved by a single pointer without dragging.",
		Prompt: `Evaluate whether drag operations have single-pointer alternatives.

Page HTML: {{.HTML}}
{{if .Screenshot}}Visual reference: [See screenshot]{{end}}

Look for draggable elements:
1. Drag-and-drop interfaces (file uploads, kanban boards, sortable lists)
2. Sliders and range inputs
3. Map panning
4. Carousel/gallery swiping
5. Resizable panels
6. Drawing/annotation tools

For each draggable element, verify an alternative exists:
- Sliders: Can also use arrow keys or direct input
- Drag-drop: Can also use buttons/menus to move items
- Maps: Can also use zoom buttons and arrow controls
- Carousels: Have prev/next buttons
- Sortable lists: Have move up/down buttons

Exemption: Functionality where dragging is essential (signature capture, freehand drawing).`,
	},

	"3.2.3": {
		ID:          "3.2.3",
		Name:        "Consistent Navigation",
		Description: "Navigational mechanisms that are repeated on multiple Web pages occur in the same relative order each time.",
		Prompt: `Evaluate whether navigation is consistent across pages.

Current page navigation: {{.HTML}}
Previous page navigation: {{.PreviousPageContext}}

Consider:
1. Are navigation items in the same relative order?
2. Are there new items added (acceptable) vs items reordered (violation)?
3. Is the primary navigation in the same location?
4. Do skip links and landmarks appear consistently?`,
	},

	"3.2.4": {
		ID:          "3.2.4",
		Name:        "Consistent Identification",
		Description: "Components that have the same functionality within a set of Web pages are identified consistently.",
		Prompt: `Evaluate whether components with the same functionality are identified consistently.

Page HTML: {{.HTML}}

Look for consistency in:
1. Search functionality - same icon, label, and position?
2. Navigation links - same labels across pages?
3. Common actions - "Submit", "Cancel", "Delete" named consistently?
4. Social media icons - same icons and alt text?
5. Login/logout - consistent labeling?
6. Shopping cart - same icon and label?

Inconsistency examples (violations):
- "Search" on one page, "Find" on another
- Magnifying glass icon sometimes labeled, sometimes not
- "Sign In" vs "Log In" vs "Login"
- Cart icon with different alt text on different pages

Note: Visual styling can vary, but functional labels and names should be consistent.`,
	},

	"3.2.6": {
		ID:          "3.2.6",
		Name:        "Consistent Help",
		Description: "If a Web page contains help mechanisms, they occur in the same relative order on each page.",
		Prompt: `Evaluate whether help mechanisms appear consistently across pages.

Page HTML: {{.HTML}}

Look for help mechanisms:
1. Contact information (phone, email, address)
2. Human contact options (chat, callback)
3. Self-help options (FAQ, knowledge base links)
4. Automated contact (chatbot)

For each help mechanism found, check:
1. Is it in a consistent location (e.g., always in footer)?
2. Does it appear in the same relative order to other help options?
3. Is the same help available on all pages that need it?

Note: Not all help mechanisms need to be on every page, but those that ARE present should be in consistent locations and order.

If this is a single-page evaluation, check that help mechanisms are clearly located and accessible.`,
	},

	"3.3.1": {
		ID:          "3.3.1",
		Name:        "Error Identification",
		Description: "If an input error is automatically detected, the item that is in error is identified and the error is described to the user in text.",
		Prompt: `Evaluate whether this error message adequately identifies the error.

Error message: {{.Element.TextContent}}
Form field: {{.SurroundingContext}}

Consider:
1. Does the error message identify WHICH field has the error?
2. Does it describe WHAT is wrong (not just "invalid input")?
3. Is the error associated with the field (aria-describedby, proximity)?
4. Would a user understand how to fix the error?`,
	},

	"3.3.3": {
		ID:          "3.3.3",
		Name:        "Error Suggestion",
		Description: "If an input error is automatically detected and suggestions for correction are known, then the suggestions are provided to the user.",
		Prompt: `Evaluate whether error suggestions are provided.

Error message: {{.Element.TextContent}}
Form field type: {{.Element.Attributes.type}}
Validation pattern: {{.Element.Attributes.pattern}}
Context: {{.SurroundingContext}}

Consider:
1. Does the error message suggest HOW to fix the problem?
2. For format errors (dates, phones), is the expected format shown?
3. For selection errors, are valid options indicated?
4. Are suggestions specific and actionable?`,
	},

	"3.3.4": {
		ID:          "3.3.4",
		Name:        "Error Prevention (Legal, Financial, Data)",
		Description: "For pages with legal commitments, financial transactions, or user data modifications, submissions are reversible, checked, or confirmed.",
		Prompt: `Evaluate error prevention for important submissions.

Page HTML: {{.HTML}}

This criterion applies to pages that:
- Cause legal commitments (agreements, contracts)
- Involve financial transactions (purchases, transfers)
- Modify/delete user-controllable data

For applicable pages, at least ONE must be true:
1. REVERSIBLE: Submissions can be undone (cancel order, edit booking)
2. CHECKED: Data is checked and user can correct errors before final submission
3. CONFIRMED: A confirmation step reviews data before committing

Look for:
- Checkout processes: Is there a review step?
- Account deletion: Is there a confirmation dialog?
- Financial forms: Can users review before submitting?
- Legal agreements: Is there a summary before accepting?

If this page doesn't involve legal/financial/data changes, mark as "Not Applicable".`,
	},

	"3.3.8": {
		ID:          "3.3.8",
		Name:        "Accessible Authentication (Minimum)",
		Description: "A cognitive function test is not required for any step in an authentication process unless alternatives are provided.",
		Prompt: `Evaluate whether authentication requires cognitive function tests.

Authentication form: {{.HTML}}
{{if .Screenshot}}See screenshot for visual verification{{end}}

Cognitive function tests include:
- CAPTCHA (image/audio puzzles)
- Math problems
- Word puzzles
- Pattern recognition
- Memory tests

Acceptable alternatives:
- Object recognition (identify photos you uploaded)
- Password managers (copy/paste allowed)
- Passkeys/biometrics
- Email/SMS codes`,
	},
}

// GetPrompt returns the prompt for a criterion, or empty if not found.
func GetPrompt(criterionID string) (CriterionPrompt, bool) {
	p, ok := CriterionPrompts[criterionID]
	return p, ok
}

// GetAllPrompts returns all criterion prompts.
func GetAllPrompts() map[string]CriterionPrompt {
	return CriterionPrompts
}
