---
name: Inkomi
colors:
  surface: '#faf8ff'
  surface-dim: '#d2d9f4'
  surface-bright: '#faf8ff'
  surface-container-lowest: '#ffffff'
  surface-container-low: '#f2f3ff'
  surface-container: '#eaedff'
  surface-container-high: '#e2e7ff'
  surface-container-highest: '#dae2fd'
  on-surface: '#131b2e'
  on-surface-variant: '#464554'
  inverse-surface: '#283044'
  inverse-on-surface: '#eef0ff'
  outline: '#767586'
  outline-variant: '#c7c4d7'
  surface-tint: '#494bd6'
  primary: '#4648d4'
  on-primary: '#ffffff'
  primary-container: '#6063ee'
  on-primary-container: '#fffbff'
  inverse-primary: '#c0c1ff'
  secondary: '#4b41e1'
  on-secondary: '#ffffff'
  secondary-container: '#645efb'
  on-secondary-container: '#fffbff'
  tertiary: '#904900'
  on-tertiary: '#ffffff'
  tertiary-container: '#b55d00'
  on-tertiary-container: '#fffbff'
  error: '#ba1a1a'
  on-error: '#ffffff'
  error-container: '#ffdad6'
  on-error-container: '#93000a'
  primary-fixed: '#e1e0ff'
  primary-fixed-dim: '#c0c1ff'
  on-primary-fixed: '#07006c'
  on-primary-fixed-variant: '#2f2ebe'
  secondary-fixed: '#e2dfff'
  secondary-fixed-dim: '#c3c0ff'
  on-secondary-fixed: '#0f0069'
  on-secondary-fixed-variant: '#3323cc'
  tertiary-fixed: '#ffdcc5'
  tertiary-fixed-dim: '#ffb783'
  on-tertiary-fixed: '#301400'
  on-tertiary-fixed-variant: '#703700'
  background: '#faf8ff'
  on-background: '#131b2e'
  surface-variant: '#dae2fd'
typography:
  display-lg:
    fontFamily: Inter
    fontSize: 32px
    fontWeight: '700'
    lineHeight: 40px
    letterSpacing: -0.02em
  display-lg-mobile:
    fontFamily: Inter
    fontSize: 28px
    fontWeight: '700'
    lineHeight: 34px
    letterSpacing: -0.02em
  headline-md:
    fontFamily: Inter
    fontSize: 20px
    fontWeight: '600'
    lineHeight: 28px
  body-lg:
    fontFamily: Inter
    fontSize: 16px
    fontWeight: '400'
    lineHeight: 24px
  body-sm:
    fontFamily: Inter
    fontSize: 14px
    fontWeight: '400'
    lineHeight: 20px
  label-caps:
    fontFamily: Inter
    fontSize: 12px
    fontWeight: '600'
    lineHeight: 16px
    letterSpacing: 0.05em
  mono-label:
    fontFamily: JetBrains Mono
    fontSize: 12px
    fontWeight: '500'
    lineHeight: 16px
rounded:
  sm: 0.25rem
  DEFAULT: 0.5rem
  md: 0.75rem
  lg: 1rem
  xl: 1.5rem
  full: 9999px
spacing:
  base: 4px
  xs: 4px
  sm: 8px
  md: 16px
  lg: 24px
  xl: 32px
  container-padding: 20px
  gutter: 12px
---

## Brand & Style
The design system is centered around "Modern Utility," a philosophy that balances the raw efficiency of a conversion tool with the immersive aesthetic of digital manga culture. The target audience includes tech-savvy readers and archival enthusiasts who value precision and clarity.

The visual style is a blend of **Minimalism** and **Tactile Modernism**. It prioritizes heavy whitespace to reduce cognitive load during complex file management tasks, while using intentional pops of "Ink" color to guide the user's eye toward primary actions. The interface should feel invisible, acting as a sophisticated frame for the manga content it processes.

## Colors
The palette is grounded in a "High-Contrast Monochrome" foundation with a "Vibrant Ink" accent.

- **Primary (Manga Ink):** A deep, vibrant indigo used for primary actions, progress bars, and active states. It mimics the richness of fresh print ink.
- **Neutral (Slate):** Used for typography and iconography to provide better readability than pure black, reducing eye strain during long sessions.
- **Surface:** A series of very light grays are used to differentiate "Cloud" storage areas from "Local" file zones.
- **Semantic:** Use a bright sky blue for cloud-sync status and a muted slate for local storage to create an immediate mental model of where files reside.

## Typography
Inter is used across the entire system for its neutral, systematic clarity. 

- **Hierarchy:** Use bold weights and negative letter-spacing for headlines to create a dense, modern look reminiscent of manga title lettering.
- **Utility:** For technical metadata (file sizes, conversion bitrates), use a monospaced alternative like JetBrains Mono to emphasize the "Utility" aspect of the tool.
- **Readability:** Body text should maintain a generous line height (1.5x) to ensure file paths and instructions are easily digestible on mobile screens.

## Layout & Spacing
This design system utilizes a **fluid grid** optimized for mobile-first interaction. 

- **Grid:** A 4-column layout for mobile with 20px side margins. 
- **Rhythm:** All spacing must be a multiple of 4px. Use 16px (md) for standard padding between elements in a card, and 24px (lg) to separate distinct sections of the workflow (e.g., File Selection vs. Conversion Settings).
- **Touch Targets:** Minimum touch target size is 44x44px. 
- **Reflow:** On larger tablet devices, the single column file list transforms into a multi-column masonry grid to showcase manga cover art.

## Elevation & Depth
Depth is signaled through **Tonal Layers** and **Ambient Shadows** rather than heavy borders.

- **Level 0 (Background):** Pure white. Used for the main canvas.
- **Level 1 (Cards/Lists):** Surface Color (#F8FAFC) with a very soft, large-radius shadow (Y: 4px, Blur: 12px, 4% Opacity of Neutral Slate).
- **Level 2 (Modals/Overlays):** Pure white with a more defined shadow (Y: 8px, Blur: 24px, 8% Opacity) to pull the element forward during critical conversion steps.
- **Interactive:** On press, elements should visually "sink" by reducing shadow spread, providing tactile feedback.

## Shapes
The shape language is friendly but structured. 

- **Base Radius:** 16px is the standard for all primary containers and cards (`rounded-lg`). 
- **Small Elements:** Chips, tags, and small buttons use a 8px radius (`rounded-md`).
- **Interactive Inputs:** Text fields use the 12px radius to strike a balance between the card and button styles.
- **Icons:** Use a consistent 2px stroke weight with rounded caps to match the UI's geometry.

## Components

### Buttons
- **Primary:** Solid Indigo background with white text. 16px height-wise padding. 
- **Secondary:** Surface-colored background with Indigo text. No border.
- **Tertiary/Ghost:** No background, Slate text. Used for "Cancel" or "Back" actions.

### Cards (Manga Items)
- Contain a 16px rounded cover image on the left.
- Use a "Cloud" or "Folder" icon in the top right to indicate source.
- Progress bars for active conversions should be 4px thick, using the Primary Ink color against a 10% opacity track.

### Form Elements
- **Input Fields:** 1px border using a light neutral tint. On focus, the border thickens to 2px Primary Ink.
- **Switches:** Use a pill-shape (60px width) with a smooth sliding animation for toggling "EPUB Optimization" settings.

### Status Indicators
- **Cloud:** Sky Blue icon with a subtle pulse animation when syncing.
- **Local:** Slate Grey icon, static.
- **Success:** Emerald green checkmark, only appearing for 2 seconds post-conversion before fading.

### Chips
- Used for manga genres or file tags (e.g., ".CBZ", ".PDF"). 
- 8px rounded corners with a subtle background tint of the text color.