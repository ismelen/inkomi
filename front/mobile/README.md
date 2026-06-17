# Converter Dashboard - React Native App

## Project Overview
This is a React Native application built with Expo and TypeScript for a file converter dashboard. The application provides a complete UI for managing folders, selecting files for conversion, configuring conversion settings, and monitoring conversion progress.

## Technology Stack
- **Framework:** React Native with Expo
- **Language:** TypeScript
- **Navigation:** React Navigation (Stack Navigator)
- **Icons:** @expo/vector-icons (Ionicons)
- **Architecture:** Component-based with reusable styled components

## Project Structure

```
ERMC-front/
├── src/
│   ├── components/          # Reusable UI components
│   │   ├── Button.tsx       # Button with variants (primary, secondary, ghost)
│   │   ├── Card.tsx         # Card container
│   │   ├── Input.tsx        # Text input with label
│   │   ├── Switch.tsx       # Toggle switch
│   │   ├── ProgressBar.tsx  # Progress indicator
│   │   ├── TabBar.tsx       # Tab navigation component
│   │   ├── FolderCard.tsx   # Folder display card
│   │   ├── FileItem.tsx     # File list item with checkbox
│   │   ├── TaskItem.tsx     # Conversion task item
│   │   └── index.ts         # Component exports
│   │
│   ├── screens/             # Application screens
│   │   ├── ConverterDashboardScreen.tsx  # Main dashboard (folders view)
│   │   ├── PendingFilesScreen.tsx        # File selection screen
│   │   ├── ConversionSettingsScreen.tsx  # Settings configuration
│   │   ├── ConversionProgressScreen.tsx  # Progress tracking
│   │   └── index.ts                      # Screen exports
│   │
│   ├── navigation/          # Navigation configuration
│   │   └── AppNavigator.tsx # Stack navigator setup
│   │
│   ├── theme/               # Theme configuration
│   │   ├── colors.ts        # Color palette
│   │   ├── fonts.ts         # Typography system
│   │   ├── spacing.ts       # Spacing, borders, shadows
│   │   └── index.ts         # Theme exports
│   │
│   ├── types/               # TypeScript types
│   │   └── index.ts         # Type definitions
│   │
│   └── App.tsx              # Main app component
│
├── App.tsx                  # Root entry point
├── app.json                 # Expo configuration
├── package.json             # Dependencies
└── tsconfig.json            # TypeScript configuration
```

## Theme System

### Colors
The app uses a carefully selected color palette:
- **Primary:** Cyan blue (#3ABAED) for actions and highlights
- **Background:** Light gray (#F5F7FA) for screens
- **Text:** Dark gray hierarchy for readability
- **Status:** Green (success/complete), Orange (progress), Red (error)

### Typography
Consistent typography scale with font sizes from 10px (caption) to 28px (large headings), using system fonts with proper weights.

### Spacing
8-point spacing system (4, 8, 12, 16, 20, 24, 32, 40, 48px) for consistent layout.

## Application Screens

### 1. Converter Dashboard (`ConverterDashboardScreen`)
- **Purpose:** Main entry point showing folder management
- **Features:**
  - Tab navigation between "Watched Folders" and "Individual Files"
  - List of active folders with status and storage information
  - "Add Folder" button (handler: TODO)
  - Settings icon (handler: TODO)
- **Navigation:** Goes to `PendingFiles` when folder is clicked

### 2. Pending Files (`PendingFilesScreen`)
- **Purpose:** File selection for conversion
- **Features:**
  - Breadcrumb navigation showing source type
  - List of files with checkboxes for multi-selection
  - "Select all" / "Deselect all" functionality
  - "Add Files" button (handler: TODO)
  - "Continue" button (navigates to Settings when files selected)
- **Navigation:** Goes to `ConversionSettings` when Continue is pressed

### 3. Conversion Settings (`ConversionSettingsScreen`)
- **Purpose:** Configure conversion parameters
- **Features:**
  - Metadata inputs (Output Filename, Author, Starting Volume Number)
  - Processing options toggles (Merge into single file, Delete after upload)
  - Output destination selection (Google Drive / Local Storage)
  - "Start Conversion" button (navigates to Progress screen)
- **Navigation:** Goes to `ConversionProgress` when conversion starts

### 4. Conversion Progress (`ConversionProgressScreen`)
- **Purpose:** Track active and completed conversions
- **Features:**
  - Active tasks section with progress bars and percentages
  - Recently completed section
  - Bottom tab navigation (Home, Convert, Library, Settings - all TODO)
  - "Clear" button for completed tasks (handler: TODO)
- **Navigation:** Has bottom tabs for future navigation

## Reusable Components

### Base Components
- **Button:** Multiple variants and sizes, supports icons and loading state
- **Card:** Container with shadow and padding
- **Input:** Text input with label and error state
- **Switch:** Toggle with label and description
- **ProgressBar:** Customizable progress indicator
- **TabBar:** Tab switcher component

### Specialized Components
- **FolderCard:** Displays folder information with status, storage, and sync time
- **FileItem:** File list item with checkbox for selection
- **TaskItem:** Shows conversion task with progress or completion status

## Data Models (TypeScript Types)

```typescript
interface Folder {
  id: string;
  name: string;
  path: string;
  status: 'monitoring' | 'idle';
  filesCount: number;
  storageUsed: string;
  lastSync?: string;
}

interface File {
  id: string;
  name: string;
  size: string;
  status: 'ready' | 'processing' | 'complete' | 'error';
  selected?: boolean;
}

interface ConversionSettings {
  outputFilename: string;
  author: string;
  startingVolumeNumber: number;
  mergeIntoSingleFile: boolean;
  deleteLocalSourceAfterUpload: boolean;
  outputDestination: 'googleDrive' | 'localStorage';
}

interface ConversionTask {
  id: string;
  name: string;
  fileName: string;
  progress: number;
  status: 'active' | 'complete' | 'ready';
  sourceFile?: string;
}
```

## TODO Handlers

The following functionality has placeholder handlers marked with `// TODO: Implement`:

### Dashboard Screen
- `handleAddFolder()` - Add new folder to watch
- `handleSettings()` - Open settings
- `handleMorePress(folderId)` - Folder actions menu

### Pending Files Screen
- `handleAddFiles()` - Add individual files

### Settings Screen
- All non-navigation interactions work with local state

### Progress Screen
- `handleClear()` - Clear completed tasks
- `handleTabPress(tab)` - Bottom tab navigation (Home, Convert, Library, Settings)

## Running the Application

### Start Development Server
```bash
npm start
# or
npx expo start
```

### Run on Specific Platform
```bash
# iOS Simulator (Mac only)
npm run ios

# Android Emulator
npm run android

# Web Browser
npm run web
```

### Build for Production
```bash
# Create production build
eas build --platform android
eas build --platform ios
```

## Future Enhancements

1. **Backend Integration**
   - Connect to actual file conversion API
   - Implement real-time progress updates via WebSocket
   - Add authentication and user management

2. **State Management**
   - Add Redux or Zustand for global state
   - Persist settings and conversion history

3. **Additional Features**
   - File preview before conversion
   - Batch operations
   - Conversion history and analytics
   - Cloud storage integration (Google Drive, Dropbox)
   - Custom conversion presets

4. **UI Improvements**
   - Dark mode support
   - Animations and transitions
   - Pull-to-refresh on lists
   - Empty states and error boundaries

5. **Testing**
   - Unit tests for components
   - Integration tests for navigation
   - E2E tests with Detox

## Dependencies

```json
{
  "@react-navigation/native": "Latest",
  "@react-navigation/stack": "Latest",
  "react-native-screens": "Latest",
  "react-native-safe-area-context": "Latest",
  "@expo/vector-icons": "Latest"
}
```

## Git Repository

The repository has been initialized with the initial commit:
```
Initial commit: React Native Expo converter dashboard with TypeScript
```

## Development Guidelines

1. **Component Creation**
   - Use functional components with TypeScript
   - Follow the established theme system
   - Maintain consistent naming conventions

2. **Styling**
   - Use StyleSheet.create for all styles
   - Reference theme values, don't hardcode
   - Follow the 8-point spacing system

3. **Navigation**
   - Use type-safe navigation with TypeScript
   - Define all routes in RootStackParamList

4. **Code Organization**
   - Keep components focused and single-purpose
   - Extract reusable logic into hooks
   - Use barrel exports (index.ts files)
