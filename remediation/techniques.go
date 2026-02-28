package remediation

// TechniqueID represents a WCAG technique identifier.
type TechniqueID string

// TechniqueType indicates whether a technique is sufficient, advisory, or a failure.
type TechniqueType string

const (
	TechniqueTypeSufficient TechniqueType = "sufficient"
	TechniqueTypeAdvisory   TechniqueType = "advisory"
	TechniqueTypeFailure    TechniqueType = "failure"
)

// TechniqueCategory represents the category of a WCAG technique.
type TechniqueCategory string

const (
	CategoryGeneral    TechniqueCategory = "general"
	CategoryHTML       TechniqueCategory = "html"
	CategoryCSS        TechniqueCategory = "css"
	CategoryScript     TechniqueCategory = "client-side-script"
	CategoryServer     TechniqueCategory = "server-side-script"
	CategorySMIL       TechniqueCategory = "smil"
	CategoryText       TechniqueCategory = "text"
	CategoryARIA       TechniqueCategory = "aria"
	CategoryPDF        TechniqueCategory = "pdf"
	CategoryFlash      TechniqueCategory = "flash"
	CategorySilverlight TechniqueCategory = "silverlight"
	CategoryFailure    TechniqueCategory = "failures"
)

// Technique contains metadata about a WCAG technique.
type Technique struct {
	ID          TechniqueID
	Category    TechniqueCategory
	Title       string
	Type        TechniqueType
	Criteria    []string // Success criteria this technique applies to
	Description string
}

// General Techniques (G)
const (
	TechniqueG1   TechniqueID = "G1"   // Adding a link at the top of each page
	TechniqueG4   TechniqueID = "G4"   // Allowing content to be paused and restarted
	TechniqueG5   TechniqueID = "G5"   // Allowing users to complete an activity without a time limit
	TechniqueG8   TechniqueID = "G8"   // Providing a movie with extended audio descriptions
	TechniqueG9   TechniqueID = "G9"   // Creating captions for live synchronized media
	TechniqueG10  TechniqueID = "G10"  // Creating components using accessible technology
	TechniqueG11  TechniqueID = "G11"  // Creating content that blinks for less than 5 seconds
	TechniqueG13  TechniqueID = "G13"  // Describing what will happen before a change
	TechniqueG14  TechniqueID = "G14"  // Ensuring text color is not sole means of conveying information
	TechniqueG15  TechniqueID = "G15"  // Using a tool to ensure content does not violate luminosity thresholds
	TechniqueG17  TechniqueID = "G17"  // Ensuring 7:1 contrast ratio for text
	TechniqueG18  TechniqueID = "G18"  // Ensuring 4.5:1 contrast ratio for text
	TechniqueG19  TechniqueID = "G19"  // Ensuring no component flashes more than 3 times per second
	TechniqueG21  TechniqueID = "G21"  // Ensuring users are not trapped in content
	TechniqueG53  TechniqueID = "G53"  // Identifying the purpose of a link using link text combined with enclosing sentence
	TechniqueG54  TechniqueID = "G54"  // Including a sign language interpreter in the video stream
	TechniqueG55  TechniqueID = "G55"  // Linking to definitions
	TechniqueG56  TechniqueID = "G56"  // Mixing audio files so speech is 20 dB above background
	TechniqueG57  TechniqueID = "G57"  // Ordering the content in a meaningful sequence
	TechniqueG58  TechniqueID = "G58"  // Placing a link to the alternative adjacent to the non-text content
	TechniqueG59  TechniqueID = "G59"  // Placing the interactive elements in an order that follows sequences
	TechniqueG60  TechniqueID = "G60"  // Playing a sound that turns off automatically within 3 seconds
	TechniqueG61  TechniqueID = "G61"  // Presenting repeated components in the same relative order
	TechniqueG62  TechniqueID = "G62"  // Providing a glossary
	TechniqueG63  TechniqueID = "G63"  // Providing a site map
	TechniqueG64  TechniqueID = "G64"  // Providing a Table of Contents
	TechniqueG65  TechniqueID = "G65"  // Providing a breadcrumb trail
	TechniqueG68  TechniqueID = "G68"  // Providing a short text alternative describing purpose of live audio-only
	TechniqueG69  TechniqueID = "G69"  // Providing an alternative for time based media
	TechniqueG70  TechniqueID = "G70"  // Providing a function to search an online dictionary
	TechniqueG71  TechniqueID = "G71"  // Providing a help link on every web page
	TechniqueG73  TechniqueID = "G73"  // Providing a long description in another location
	TechniqueG74  TechniqueID = "G74"  // Providing a long description in text near the non-text content
	TechniqueG75  TechniqueID = "G75"  // Providing a mechanism to postpone any updating of content
	TechniqueG76  TechniqueID = "G76"  // Providing a mechanism to request an update of the content
	TechniqueG78  TechniqueID = "G78"  // Providing a second, user-selectable, audio track with audio descriptions
	TechniqueG79  TechniqueID = "G79"  // Providing a spoken version of the text
	TechniqueG80  TechniqueID = "G80"  // Providing a submit button to initiate a change of context
	TechniqueG81  TechniqueID = "G81"  // Providing a synchronized video of the sign language interpreter
	TechniqueG82  TechniqueID = "G82"  // Providing a text alternative that identifies purpose of non-text content
	TechniqueG83  TechniqueID = "G83"  // Providing text descriptions to identify required fields
	TechniqueG84  TechniqueID = "G84"  // Providing a text description when user provides information not in allowed values
	TechniqueG85  TechniqueID = "G85"  // Providing a text description when user input is outside required format
	TechniqueG86  TechniqueID = "G86"  // Providing a text summary that can be understood by people with lower reading ability
	TechniqueG87  TechniqueID = "G87"  // Providing closed captions
	TechniqueG88  TechniqueID = "G88"  // Providing descriptive titles for web pages
	TechniqueG89  TechniqueID = "G89"  // Providing expected data format and example
	TechniqueG90  TechniqueID = "G90"  // Providing keyboard-triggered event handlers
	TechniqueG91  TechniqueID = "G91"  // Providing link text that describes the purpose of a link
	TechniqueG92  TechniqueID = "G92"  // Providing long description for non-text content via role
	TechniqueG93  TechniqueID = "G93"  // Providing open captions
	TechniqueG94  TechniqueID = "G94"  // Providing short text alternative for non-text content
	TechniqueG95  TechniqueID = "G95"  // Providing short text alternatives for brief description
	TechniqueG96  TechniqueID = "G96"  // Providing textual identification of items that use sensory information
	TechniqueG97  TechniqueID = "G97"  // Providing the first use of an abbreviation immediately followed by expansion
	TechniqueG98  TechniqueID = "G98"  // Providing ability for user to review and correct answers
	TechniqueG99  TechniqueID = "G99"  // Providing ability to recover deleted information
	TechniqueG100 TechniqueID = "G100" // Providing short text alternative for non-text content serving sensory experience
	TechniqueG101 TechniqueID = "G101" // Providing the definition of a word or phrase used in an unusual way
	TechniqueG102 TechniqueID = "G102" // Providing the expansion or explanation of an abbreviation
	TechniqueG103 TechniqueID = "G103" // Providing visual illustrations, pictures, and symbols
	TechniqueG105 TechniqueID = "G105" // Saving data so it can be used after reauthentication
	TechniqueG107 TechniqueID = "G107" // Using "activate" rather than "focus" as a trigger
	TechniqueG108 TechniqueID = "G108" // Using markup features to expose the name and role
	TechniqueG110 TechniqueID = "G110" // Using an instant client-side redirect
	TechniqueG111 TechniqueID = "G111" // Using color and pattern
	TechniqueG112 TechniqueID = "G112" // Using inline definitions
	TechniqueG115 TechniqueID = "G115" // Using semantic elements to mark up structure
	TechniqueG117 TechniqueID = "G117" // Using text to convey information conveyed by color variations
	TechniqueG120 TechniqueID = "G120" // Providing the pronunciation immediately following the word
	TechniqueG121 TechniqueID = "G121" // Linking to pronunciations
	TechniqueG123 TechniqueID = "G123" // Adding a link at the beginning of a block to go to the end
	TechniqueG124 TechniqueID = "G124" // Adding links at the top of the page to each area of the content
	TechniqueG125 TechniqueID = "G125" // Providing links to navigate to related web pages
	TechniqueG126 TechniqueID = "G126" // Providing a list of links to all other web pages
	TechniqueG127 TechniqueID = "G127" // Identifying a web page's relationship to a larger collection
	TechniqueG128 TechniqueID = "G128" // Indicating current location within navigation bars
	TechniqueG130 TechniqueID = "G130" // Providing descriptive headings
	TechniqueG131 TechniqueID = "G131" // Providing descriptive labels
	TechniqueG133 TechniqueID = "G133" // Providing a checkbox on the first page of a multipart form
	TechniqueG134 TechniqueID = "G134" // Validating web pages
	TechniqueG135 TechniqueID = "G135" // Using the accessibility API features
	TechniqueG136 TechniqueID = "G136" // Providing a link at the beginning of a nonconforming page
	TechniqueG138 TechniqueID = "G138" // Using semantic markup whenever color cues are used
	TechniqueG139 TechniqueID = "G139" // Creating a mechanism to skip blocks of content
	TechniqueG140 TechniqueID = "G140" // Separating information and structure from presentation
	TechniqueG141 TechniqueID = "G141" // Organizing a page using headings
	TechniqueG142 TechniqueID = "G142" // Using a technology that has commonly available user agents supporting zoom
	TechniqueG143 TechniqueID = "G143" // Providing a text alternative describing purpose of CAPTCHA
	TechniqueG144 TechniqueID = "G144" // Ensuring web pages contain another CAPTCHA serving same purpose
	TechniqueG145 TechniqueID = "G145" // Ensuring 3:1 contrast ratio for larger text
	TechniqueG146 TechniqueID = "G146" // Using liquid layout
	TechniqueG148 TechniqueID = "G148" // Not specifying background color or specifying foreground color
	TechniqueG149 TechniqueID = "G149" // Using user interface components highlighted when they receive focus
	TechniqueG150 TechniqueID = "G150" // Providing text-based alternatives for live audio-only content
	TechniqueG151 TechniqueID = "G151" // Providing a link to a text transcript
	TechniqueG152 TechniqueID = "G152" // Setting animated gif images to stop blinking after n cycles
	TechniqueG153 TechniqueID = "G153" // Making text easier to read
	TechniqueG155 TechniqueID = "G155" // Providing a checkbox in addition to a submit button
	TechniqueG156 TechniqueID = "G156" // Using a technology that has commonly available user agents
	TechniqueG157 TechniqueID = "G157" // Incorporating a live audio captioning service into a web page
	TechniqueG158 TechniqueID = "G158" // Providing an alternative for time-based media for audio-only content
	TechniqueG159 TechniqueID = "G159" // Providing an alternative for time-based media for video-only content
	TechniqueG160 TechniqueID = "G160" // Providing sign language versions of information
	TechniqueG161 TechniqueID = "G161" // Providing a search function to help users find content
	TechniqueG162 TechniqueID = "G162" // Positioning labels to maximize predictability of relationships
	TechniqueG163 TechniqueID = "G163" // Using standard diacritical marks that can be turned off
	TechniqueG164 TechniqueID = "G164" // Providing a stated time within which an online request will be confirmed
	TechniqueG165 TechniqueID = "G165" // Using the default focus indicator for the platform
	TechniqueG166 TechniqueID = "G166" // Providing audio that describes important video content
	TechniqueG167 TechniqueID = "G167" // Using an adjacent button to label purpose of a field
	TechniqueG168 TechniqueID = "G168" // Requesting confirmation to continue with selected action
	TechniqueG169 TechniqueID = "G169" // Aligning text on only one side
	TechniqueG170 TechniqueID = "G170" // Providing a control near the top of the page to turn off sounds
	TechniqueG171 TechniqueID = "G171" // Playing sounds only on user request
	TechniqueG172 TechniqueID = "G172" // Providing a mechanism to remove full justification of text
	TechniqueG173 TechniqueID = "G173" // Providing a version of a movie with audio descriptions
	TechniqueG174 TechniqueID = "G174" // Providing a control with sufficient contrast to toggle to sufficient contrast
	TechniqueG175 TechniqueID = "G175" // Providing a multi color selection tool on the page
	TechniqueG176 TechniqueID = "G176" // Keeping the flashing area small enough
	TechniqueG177 TechniqueID = "G177" // Providing suggested correction text
	TechniqueG178 TechniqueID = "G178" // Providing controls on the web page to incrementally change text size
	TechniqueG179 TechniqueID = "G179" // Ensuring text and background resize when text is resized
	TechniqueG180 TechniqueID = "G180" // Providing the user with a means to set the time limit to 10 times the default
	TechniqueG181 TechniqueID = "G181" // Encoding user data as hidden or encrypted in a re-authorization page
	TechniqueG182 TechniqueID = "G182" // Ensuring that additional visual cues are available when text color differences are used
	TechniqueG183 TechniqueID = "G183" // Using 3:1 contrast ratio with surrounding text and providing additional visual cues
	TechniqueG184 TechniqueID = "G184" // Providing text instructions at the beginning of a form
	TechniqueG185 TechniqueID = "G185" // Linking to all pages from the home page
	TechniqueG186 TechniqueID = "G186" // Using a control in the web page to stop content that blinks
	TechniqueG187 TechniqueID = "G187" // Using a technology to include blinking content that can be turned off via user agent
	TechniqueG188 TechniqueID = "G188" // Providing a button on the page to increase line spaces and paragraph spaces
	TechniqueG189 TechniqueID = "G189" // Providing a control near the beginning of the page to change the link text
	TechniqueG190 TechniqueID = "G190" // Providing a link adjacent to or associated with a non-conforming object
	TechniqueG191 TechniqueID = "G191" // Providing a link, button, or other mechanism that reloads the page without blinking
	TechniqueG192 TechniqueID = "G192" // Fully conforming to specifications
	TechniqueG193 TechniqueID = "G193" // Providing help by an assistant in the web page
	TechniqueG194 TechniqueID = "G194" // Providing spell checking and suggestions for text input
	TechniqueG195 TechniqueID = "G195" // Using an author-supplied, visible focus indicator
	TechniqueG196 TechniqueID = "G196" // Using a text alternative on one item within a group
	TechniqueG197 TechniqueID = "G197" // Using labels, names, and text alternatives consistently
	TechniqueG198 TechniqueID = "G198" // Providing a way for the user to turn the time limit off
	TechniqueG199 TechniqueID = "G199" // Providing success feedback when data is submitted successfully
	TechniqueG200 TechniqueID = "G200" // Opening new windows and tabs from a link only when necessary
	TechniqueG201 TechniqueID = "G201" // Giving users advanced warning when opening a new window
	TechniqueG202 TechniqueID = "G202" // Ensuring keyboard control for all functionality
	TechniqueG203 TechniqueID = "G203" // Using a static text alternative to describe a talking head video
	TechniqueG204 TechniqueID = "G204" // Not interfering with the user agent's reflow of text
	TechniqueG205 TechniqueID = "G205" // Including a text cue whenever color cue is used
	TechniqueG206 TechniqueID = "G206" // Providing options within the content to switch to a layout that does not require scrolling
	TechniqueG207 TechniqueID = "G207" // Ensuring that a contrast ratio of 3:1 is provided for icons
	TechniqueG208 TechniqueID = "G208" // Including the text of the visible label as part of the accessible name
	TechniqueG209 TechniqueID = "G209" // Provide sufficient contrast at the boundaries between adjoining colors
	TechniqueG210 TechniqueID = "G210" // Ensuring that drag-and-drop actions can be cancelled
	TechniqueG211 TechniqueID = "G211" // Matching the accessible name to the visible label
	TechniqueG212 TechniqueID = "G212" // Using native controls to ensure functionality is triggered on the up-event
	TechniqueG213 TechniqueID = "G213" // Provide conventional controls and an application setting for motion activated input
	TechniqueG214 TechniqueID = "G214" // Using a control to allow access to content in different orientations
	TechniqueG215 TechniqueID = "G215" // Providing controls to achieve the same result as path based or multipoint gestures
	TechniqueG216 TechniqueID = "G216" // Providing single point activation for a control slider
	TechniqueG217 TechniqueID = "G217" // Providing a mechanism to allow users to remap or turn off character key shortcuts
	TechniqueG218 TechniqueID = "G218" // Email link authentication
	TechniqueG219 TechniqueID = "G219" // Ensuring that an alternative is available for dragging movements
	TechniqueG220 TechniqueID = "G220" // Provide a contact-us link in a consistent location
	TechniqueG221 TechniqueID = "G221" // Provide data from a previous step in a process
)

// HTML Techniques (H)
const (
	TechniqueH2   TechniqueID = "H2"   // Combining adjacent image and text links
	TechniqueH4   TechniqueID = "H4"   // Creating a logical tab order through links, form controls, and objects
	TechniqueH24  TechniqueID = "H24"  // Providing text alternatives for image map areas
	TechniqueH25  TechniqueID = "H25"  // Providing a title using the title element
	TechniqueH28  TechniqueID = "H28"  // Providing definitions for abbreviations by using abbr element
	TechniqueH30  TechniqueID = "H30"  // Providing link text that describes the purpose of a link
	TechniqueH32  TechniqueID = "H32"  // Providing submit buttons
	TechniqueH33  TechniqueID = "H33"  // Supplementing link text with the title attribute
	TechniqueH34  TechniqueID = "H34"  // Using a Unicode right-to-left mark or left-to-right mark
	TechniqueH35  TechniqueID = "H35"  // Providing text alternatives on applet elements
	TechniqueH36  TechniqueID = "H36"  // Using alt attributes on images used as submit buttons
	TechniqueH37  TechniqueID = "H37"  // Using alt attributes on img elements
	TechniqueH39  TechniqueID = "H39"  // Using caption elements to associate data table captions with data tables
	TechniqueH40  TechniqueID = "H40"  // Using description lists
	TechniqueH42  TechniqueID = "H42"  // Using h1-h6 to identify headings
	TechniqueH43  TechniqueID = "H43"  // Using id and headers attributes to associate data cells with header cells
	TechniqueH44  TechniqueID = "H44"  // Using label elements to associate text labels with form controls
	TechniqueH45  TechniqueID = "H45"  // Using longdesc
	TechniqueH46  TechniqueID = "H46"  // Using noembed with embed
	TechniqueH48  TechniqueID = "H48"  // Using ol, ul and dl for lists or groups of links
	TechniqueH49  TechniqueID = "H49"  // Using semantic markup to mark emphasized or special text
	TechniqueH50  TechniqueID = "H50"  // Using structural elements to group links
	TechniqueH51  TechniqueID = "H51"  // Using table markup to present tabular information
	TechniqueH53  TechniqueID = "H53"  // Using the body of the object element
	TechniqueH54  TechniqueID = "H54"  // Using the dfn element to identify the defining instance of a word
	TechniqueH56  TechniqueID = "H56"  // Using the dir attribute on an inline element to resolve directionality
	TechniqueH57  TechniqueID = "H57"  // Using the language attribute on the HTML element
	TechniqueH58  TechniqueID = "H58"  // Using language attributes to identify changes in the human language
	TechniqueH59  TechniqueID = "H59"  // Using the link element and navigation tools
	TechniqueH60  TechniqueID = "H60"  // Using the link element to link to a glossary
	TechniqueH62  TechniqueID = "H62"  // Using the ruby element
	TechniqueH63  TechniqueID = "H63"  // Using the scope attribute to associate header cells and data cells
	TechniqueH64  TechniqueID = "H64"  // Using the title attribute of the iframe and frame elements
	TechniqueH65  TechniqueID = "H65"  // Using the title attribute to identify form controls
	TechniqueH67  TechniqueID = "H67"  // Using null alt text and no title attribute for decorative images
	TechniqueH69  TechniqueID = "H69"  // Providing heading elements at the beginning of each section
	TechniqueH70  TechniqueID = "H70"  // Using frame elements to group blocks of repeated material
	TechniqueH71  TechniqueID = "H71"  // Providing a description for groups of form controls using fieldset and legend
	TechniqueH73  TechniqueID = "H73"  // Using the summary attribute of the table element
	TechniqueH74  TechniqueID = "H74"  // Ensuring that opening and closing tags are used according to specification
	TechniqueH75  TechniqueID = "H75"  // Ensuring that web pages are well-formed
	TechniqueH76  TechniqueID = "H76"  // Using meta refresh to create an instant client-side redirect
	TechniqueH77  TechniqueID = "H77"  // Identifying the purpose of a link using link text combined with its enclosing list item
	TechniqueH78  TechniqueID = "H78"  // Identifying the purpose of a link using link text combined with its enclosing paragraph
	TechniqueH79  TechniqueID = "H79"  // Identifying the purpose of a link using link text combined with preceding heading
	TechniqueH80  TechniqueID = "H80"  // Identifying the purpose of a link using link text combined with preceding heading
	TechniqueH81  TechniqueID = "H81"  // Identifying the purpose of a link in a nested list using link text combined with parent list
	TechniqueH82  TechniqueID = "H82"  // Using a text alternative on an object element for fallback purposes
	TechniqueH83  TechniqueID = "H83"  // Using the target attribute to open a new window and indicating this in link text
	TechniqueH84  TechniqueID = "H84"  // Using a button with a select element to perform an action
	TechniqueH85  TechniqueID = "H85"  // Using optgroup to group option elements inside a select
	TechniqueH86  TechniqueID = "H86"  // Providing text alternatives for ASCII art, emoticons, and leetspeak
	TechniqueH88  TechniqueID = "H88"  // Using HTML according to spec
	TechniqueH89  TechniqueID = "H89"  // Using the title attribute to provide context-sensitive help
	TechniqueH90  TechniqueID = "H90"  // Indicating required form controls using label or legend
	TechniqueH91  TechniqueID = "H91"  // Using HTML form controls and links
	TechniqueH93  TechniqueID = "H93"  // Ensuring that id attributes are unique on a web page
	TechniqueH94  TechniqueID = "H94"  // Ensuring that elements do not contain duplicate attributes
	TechniqueH95  TechniqueID = "H95"  // Using the track element to provide captions
	TechniqueH96  TechniqueID = "H96"  // Using the track element to provide audio descriptions
	TechniqueH97  TechniqueID = "H97"  // Grouping related links using the nav element
	TechniqueH98  TechniqueID = "H98"  // Using HTML 5.2 autocomplete attributes
	TechniqueH99  TechniqueID = "H99"  // Provide a page-selection mechanism
	TechniqueH100 TechniqueID = "H100" // Validating input
	TechniqueH101 TechniqueID = "H101" // Using semantic HTML elements to identify regions of a page
)

// CSS Techniques (C)
const (
	TechniqueC6  TechniqueID = "C6"  // Positioning content based on structural markup
	TechniqueC7  TechniqueID = "C7"  // Using CSS to hide a portion of the link text
	TechniqueC8  TechniqueID = "C8"  // Using CSS letter-spacing to control spacing within a word
	TechniqueC9  TechniqueID = "C9"  // Using CSS to include decorative images
	TechniqueC12 TechniqueID = "C12" // Using percent for font sizes
	TechniqueC13 TechniqueID = "C13" // Using named font sizes
	TechniqueC14 TechniqueID = "C14" // Using em units for font sizes
	TechniqueC15 TechniqueID = "C15" // Using CSS to change the presentation of a user interface component
	TechniqueC17 TechniqueID = "C17" // Scaling form elements which contain text
	TechniqueC18 TechniqueID = "C18" // Using CSS margin and padding rules instead of spacer images
	TechniqueC19 TechniqueID = "C19" // Specifying alignment either to the left or right in CSS
	TechniqueC20 TechniqueID = "C20" // Using relative measurements to set column widths
	TechniqueC21 TechniqueID = "C21" // Specifying line spacing in CSS
	TechniqueC22 TechniqueID = "C22" // Using CSS to control visual presentation of text
	TechniqueC23 TechniqueID = "C23" // Specifying text and background colors of secondary content
	TechniqueC24 TechniqueID = "C24" // Using percentage values in CSS for container sizes
	TechniqueC25 TechniqueID = "C25" // Specifying borders and layout in CSS to delineate areas
	TechniqueC27 TechniqueID = "C27" // Making the DOM order match the visual order
	TechniqueC28 TechniqueID = "C28" // Specifying the size of text containers using em units
	TechniqueC29 TechniqueID = "C29" // Using a style switcher to provide a conforming alternate version
	TechniqueC30 TechniqueID = "C30" // Using CSS to replace text with images and providing user interface controls
	TechniqueC31 TechniqueID = "C31" // Using CSS Flexbox to reflow content
	TechniqueC32 TechniqueID = "C32" // Using media queries and grid CSS to reflow columns
	TechniqueC33 TechniqueID = "C33" // Allowing for reflow with long URLs and strings of text
	TechniqueC34 TechniqueID = "C34" // Using media queries to un-fix sticky headers
	TechniqueC35 TechniqueID = "C35" // Allowing for text spacing without wrapping
	TechniqueC36 TechniqueID = "C36" // Allowing for text spacing override
	TechniqueC37 TechniqueID = "C37" // Using CSS max-width and height to fit images
	TechniqueC38 TechniqueID = "C38" // Using CSS width, max-width and flexbox to fit labels and inputs
	TechniqueC39 TechniqueID = "C39" // Using the CSS reduce-motion query to prevent motion
	TechniqueC40 TechniqueID = "C40" // Creating a two-color focus indicator to ensure sufficient contrast
	TechniqueC41 TechniqueID = "C41" // Creating a strong focus indicator within the component
	TechniqueC42 TechniqueID = "C42" // Using min-height and min-width to ensure sufficient target spacing
	TechniqueC43 TechniqueID = "C43" // Using CSS scroll-padding to un-obscure content
	TechniqueC44 TechniqueID = "C44" // Using CSS margin to visually increase spacing between targets
	TechniqueC45 TechniqueID = "C45" // Using CSS :focus-visible to provide keyboard focus indication
)

// ARIA Techniques
const (
	TechniqueARIA1  TechniqueID = "ARIA1"  // Using the aria-describedby property to provide a descriptive label
	TechniqueARIA2  TechniqueID = "ARIA2"  // Identifying a required field with the aria-required property
	TechniqueARIA4  TechniqueID = "ARIA4"  // Using a WAI-ARIA role to expose the role of a user interface component
	TechniqueARIA5  TechniqueID = "ARIA5"  // Using WAI-ARIA state and property attributes to expose component state
	TechniqueARIA6  TechniqueID = "ARIA6"  // Using aria-label to provide labels for objects
	TechniqueARIA7  TechniqueID = "ARIA7"  // Using aria-labelledby for link purpose
	TechniqueARIA8  TechniqueID = "ARIA8"  // Using aria-label for link purpose
	TechniqueARIA9  TechniqueID = "ARIA9"  // Using aria-labelledby to concatenate a label from several text nodes
	TechniqueARIA10 TechniqueID = "ARIA10" // Using aria-labelledby to provide a text alternative for non-text content
	TechniqueARIA11 TechniqueID = "ARIA11" // Using ARIA landmarks to identify regions of a page
	TechniqueARIA12 TechniqueID = "ARIA12" // Using role=heading to identify headings
	TechniqueARIA13 TechniqueID = "ARIA13" // Using aria-labelledby to name regions and landmarks
	TechniqueARIA14 TechniqueID = "ARIA14" // Using aria-label to provide an invisible label
	TechniqueARIA15 TechniqueID = "ARIA15" // Using aria-describedby to provide descriptions of images
	TechniqueARIA16 TechniqueID = "ARIA16" // Using aria-labelledby to provide a name for user interface controls
	TechniqueARIA17 TechniqueID = "ARIA17" // Using grouping roles to identify related form controls
	TechniqueARIA18 TechniqueID = "ARIA18" // Using aria-alertdialog to identify errors
	TechniqueARIA19 TechniqueID = "ARIA19" // Using ARIA role=alert or live regions to identify errors
	TechniqueARIA20 TechniqueID = "ARIA20" // Using the region role to identify a region of the page
	TechniqueARIA21 TechniqueID = "ARIA21" // Using aria-invalid to indicate an error field
	TechniqueARIA22 TechniqueID = "ARIA22" // Using role=status to present status messages
	TechniqueARIA23 TechniqueID = "ARIA23" // Using role=log to identify sequential information updates
	TechniqueARIA24 TechniqueID = "ARIA24" // Semantically identifying a font icon with role="img"
)

// Scripting Techniques (SCR)
const (
	TechniqueSCR1  TechniqueID = "SCR1"  // Allowing the user to extend the default time limit
	TechniqueSCR2  TechniqueID = "SCR2"  // Using redundant keyboard and mouse event handlers
	TechniqueSCR14 TechniqueID = "SCR14" // Using scripts to make nonessential alerts optional
	TechniqueSCR16 TechniqueID = "SCR16" // Providing a script that warns the user a time limit is about to expire
	TechniqueSCR18 TechniqueID = "SCR18" // Providing client-side validation and alert
	TechniqueSCR19 TechniqueID = "SCR19" // Using an onchange event on a select element without causing a change of context
	TechniqueSCR20 TechniqueID = "SCR20" // Using both keyboard and other device-specific functions
	TechniqueSCR21 TechniqueID = "SCR21" // Using functions of the Document Object Model (DOM) to add content
	TechniqueSCR22 TechniqueID = "SCR22" // Using scripts to control blinking and stop it in five seconds or less
	TechniqueSCR24 TechniqueID = "SCR24" // Using progressive enhancement to open new windows on user request
	TechniqueSCR26 TechniqueID = "SCR26" // Inserting dynamic content into the Document Object Model
	TechniqueSCR27 TechniqueID = "SCR27" // Reordering page sections using the Document Object Model
	TechniqueSCR28 TechniqueID = "SCR28" // Using an expandable and collapsible menu to bypass blocks
	TechniqueSCR29 TechniqueID = "SCR29" // Adding keyboard-accessible actions to static HTML elements
	TechniqueSCR30 TechniqueID = "SCR30" // Using scripts to change the link text
	TechniqueSCR31 TechniqueID = "SCR31" // Using script to change the background color or border of the element with focus
	TechniqueSCR32 TechniqueID = "SCR32" // Providing client-side validation and adding error text via the DOM
	TechniqueSCR33 TechniqueID = "SCR33" // Using script to scroll content and providing a mechanism to pause it
	TechniqueSCR34 TechniqueID = "SCR34" // Calculating size and position in a way that scales with text size
	TechniqueSCR35 TechniqueID = "SCR35" // Making actions keyboard accessible by using the onclick event
	TechniqueSCR36 TechniqueID = "SCR36" // Providing a mechanism to allow users to display moving, scrolling, or auto-updating text
	TechniqueSCR37 TechniqueID = "SCR37" // Creating custom dialogs in a device independent way
	TechniqueSCR38 TechniqueID = "SCR38" // Creating a conforming alternate version for a web page designed with progressive enhancement
	TechniqueSCR39 TechniqueID = "SCR39" // Making content on focus or hover hoverable, dismissible, and persistent
)

// Failure Techniques (F)
const (
	TechniqueF1   TechniqueID = "F1"   // Failure due to changing context on input
	TechniqueF2   TechniqueID = "F2"   // Failure due to using CSS to include images conveying important information
	TechniqueF3   TechniqueID = "F3"   // Failure due to using CSS to include images to convey important information
	TechniqueF4   TechniqueID = "F4"   // Failure due to allowing a time limit to be reset
	TechniqueF7   TechniqueID = "F7"   // Failure due to an object or applet that only has keyboard trap
	TechniqueF8   TechniqueID = "F8"   // Failure due to a visual lists not being coded as lists
	TechniqueF9   TechniqueID = "F9"   // Failure due to changing context on focus
	TechniqueF10  TechniqueID = "F10"  // Failure due to combining multiple content formats
	TechniqueF12  TechniqueID = "F12"  // Failure due to same resource having two names
	TechniqueF13  TechniqueID = "F13"  // Failure due to image text not included in alt
	TechniqueF14  TechniqueID = "F14"  // Failure due to using units that do not support adaptation
	TechniqueF15  TechniqueID = "F15"  // Failure due to implementing custom controls that do not use platform API
	TechniqueF16  TechniqueID = "F16"  // Failure due to including scrolling content
	TechniqueF17  TechniqueID = "F17"  // Failure due to missing keyboard access
	TechniqueF19  TechniqueID = "F19"  // Failure due to not providing data in tables using headers
	TechniqueF20  TechniqueID = "F20"  // Failure due to not updating text alternatives
	TechniqueF22  TechniqueID = "F22"  // Failure due to changes in content not available to AT
	TechniqueF23  TechniqueID = "F23"  // Failure due to using background images to convey information
	TechniqueF24  TechniqueID = "F24"  // Failure due to specifying only foreground or background color
	TechniqueF25  TechniqueID = "F25"  // Failure due to presenting title as the only navigation
	TechniqueF26  TechniqueID = "F26"  // Failure due to using images to represent text
	TechniqueF30  TechniqueID = "F30"  // Failure due to using text alternatives that are not alternatives
	TechniqueF31  TechniqueID = "F31"  // Failure due to using script to remove focus when it is received
	TechniqueF32  TechniqueID = "F32"  // Failure due to using text-only substitutes for character key shortcuts
	TechniqueF33  TechniqueID = "F33"  // Failure due to providing multiple formats without links
	TechniqueF34  TechniqueID = "F34"  // Failure due to using characters to format text
	TechniqueF36  TechniqueID = "F36"  // Failure due to form automatically submitting
	TechniqueF37  TechniqueID = "F37"  // Failure due to using image buttons with no alternative text
	TechniqueF38  TechniqueID = "F38"  // Failure due to not marking up decorative images properly
	TechniqueF39  TechniqueID = "F39"  // Failure due to text alternatives not providing same information
	TechniqueF40  TechniqueID = "F40"  // Failure due to using meta redirect with time limit
	TechniqueF41  TechniqueID = "F41"  // Failure due to using meta refresh to reload the page
	TechniqueF42  TechniqueID = "F42"  // Failure due to using script handlers on emulated links
	TechniqueF43  TechniqueID = "F43"  // Failure due to using structural markup for presentation
	TechniqueF44  TechniqueID = "F44"  // Failure due to using role="presentation" incorrectly
	TechniqueF46  TechniqueID = "F46"  // Failure due to using th elements, scope attributes, headers/id incorrectly
	TechniqueF47  TechniqueID = "F47"  // Failure due to using blink element
	TechniqueF48  TechniqueID = "F48"  // Failure due to using pre to markup tabular info
	TechniqueF49  TechniqueID = "F49"  // Failure due to using HTML layout tables that do not linearize
	TechniqueF50  TechniqueID = "F50"  // Failure due to the inability to pause moving info
	TechniqueF52  TechniqueID = "F52"  // Failure due to opening new window when it could have been avoided
	TechniqueF54  TechniqueID = "F54"  // Failure due to using table only to position content
	TechniqueF55  TechniqueID = "F55"  // Failure due to using script to remove focus
	TechniqueF58  TechniqueID = "F58"  // Failure due to using script to update content in a way that is not visible
	TechniqueF59  TechniqueID = "F59"  // Failure due to using script to provide link text
	TechniqueF60  TechniqueID = "F60"  // Failure due to opening new window when the user clicks
	TechniqueF61  TechniqueID = "F61"  // Failure due to updating content through a mechanism that does not allow focus
	TechniqueF63  TechniqueID = "F63"  // Failure due to providing captions that do not identify audio info
	TechniqueF65  TechniqueID = "F65"  // Failure due to omitting alt on img elements
	TechniqueF66  TechniqueID = "F66"  // Failure due to presenting navigation links only as images
	TechniqueF67  TechniqueID = "F67"  // Failure due to providing long descriptions for non-text content
	TechniqueF68  TechniqueID = "F68"  // Failure due to association of label and form control not being identified
	TechniqueF69  TechniqueID = "F69"  // Failure due to the link text not identifying the purpose
	TechniqueF70  TechniqueID = "F70"  // Failure due to using alt text that does not describe the image
	TechniqueF71  TechniqueID = "F71"  // Failure due to providing an empty alt attribute
	TechniqueF72  TechniqueID = "F72"  // Failure due to using ASCII art without providing an alternative
	TechniqueF73  TechniqueID = "F73"  // Failure due to creating links that are not visually evident without color
	TechniqueF74  TechniqueID = "F74"  // Failure due to not having sufficient visual distinction between active states
	TechniqueF75  TechniqueID = "F75"  // Failure due to not having sufficient contrast for same word
	TechniqueF77  TechniqueID = "F77"  // Failure due to duplicating id attribute values
	TechniqueF78  TechniqueID = "F78"  // Failure due to styling element outlines and borders removing default focus
	TechniqueF79  TechniqueID = "F79"  // Failure due to focus indicator not providing enough contrast
	TechniqueF80  TechniqueID = "F80"  // Failure due to not using native mechanism for accessible name
	TechniqueF81  TechniqueID = "F81"  // Failure due to using position-dependent info for determining name
	TechniqueF82  TechniqueID = "F82"  // Failure due to using same name for different link destinations
	TechniqueF83  TechniqueID = "F83"  // Failure due to using images of text where text would have been more accessible
	TechniqueF84  TechniqueID = "F84"  // Failure due to using a non-specific link such as click here
	TechniqueF85  TechniqueID = "F85"  // Failure due to using dialogs that are not accessible to assistive tech
	TechniqueF86  TechniqueID = "F86"  // Failure due to providing accessible names without real text content
	TechniqueF87  TechniqueID = "F87"  // Failure due to inserting content using :before and :after
	TechniqueF88  TechniqueID = "F88"  // Failure due to using scroll position dependent content
	TechniqueF89  TechniqueID = "F89"  // Failure due to using null alt on images that convey information
	TechniqueF90  TechniqueID = "F90"  // Failure due to incorrectly associating table headers and content cells
	TechniqueF91  TechniqueID = "F91"  // Failure due to not correctly marking up table headers
	TechniqueF92  TechniqueID = "F92"  // Failure due to using role attribute on elements with semantic meaning
	TechniqueF93  TechniqueID = "F93"  // Failure due to lack of accessible authentication
	TechniqueF94  TechniqueID = "F94"  // Failure due to incorrect use of viewport units
	TechniqueF95  TechniqueID = "F95"  // Failure due to providing accessible names without visible text
	TechniqueF96  TechniqueID = "F96"  // Failure due to accessible name not matching visible label
	TechniqueF97  TechniqueID = "F97"  // Failure due to content disappearing and reappearing
	TechniqueF98  TechniqueID = "F98"  // Failure due to interactions limited to touch
	TechniqueF99  TechniqueID = "F99"  // Failure due to using only gesture or path
	TechniqueF100 TechniqueID = "F100" // Failure due to submitting without confirmation
	TechniqueF101 TechniqueID = "F101" // Failure due to using fixed width containers
	TechniqueF102 TechniqueID = "F102" // Failure due to content being cut off or overlapping
	TechniqueF103 TechniqueID = "F103" // Failure due to not providing authentication alternatives
	TechniqueF104 TechniqueID = "F104" // Failure due to small touch targets
	TechniqueF105 TechniqueID = "F105" // Failure due to character key shortcuts activated without modifier
	TechniqueF106 TechniqueID = "F106" // Failure due to dismissible content blocking interaction
	TechniqueF107 TechniqueID = "F107" // Failure due to using hover-only activation
)
