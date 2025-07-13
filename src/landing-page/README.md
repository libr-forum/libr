# Libr Landing Page

A modern, responsive landing page for Libr - the censorship-resistant social network.

## Features

- **Modern Design**: Clean, professional interface following current web design trends
- **Responsive Layout**: Optimized for all device sizes from mobile to desktop
- **Dark/Light Mode**: Automatic theme detection with manual toggle
- **Smooth Animations**: Framer Motion powered animations and transitions
- **Performance Optimized**: Fast loading times and smooth scrolling
- **SEO Friendly**: Proper meta tags and semantic HTML structure

## Design Principles

Based on landing page best practices research, this page includes:

- **Clear Value Proposition**: Immediately communicates Libr's core benefits
- **Social Proof**: Community stats and user testimonials
- **Progressive Disclosure**: Information revealed as users scroll
- **Strong CTAs**: Clear call-to-action buttons strategically placed
- **Trust Signals**: Open source badges, security features highlighted
- **Mobile-First**: Responsive design optimized for all devices

## Tech Stack

- **React 18** - Modern React with TypeScript
- **Vite** - Fast development and build tooling
- **Tailwind CSS** - Utility-first CSS framework
- **Framer Motion** - Animation library for smooth interactions
- **Lucide React** - Beautiful, consistent icons
- **TypeScript** - Type safety and better developer experience

## Theme System

The landing page uses the same color palette as the main Libr application:

### Light Mode
- Primary: `#FDFCF7`
- Secondary: `#304a78` 
- Accent 1: `#60B3F0`
- Accent 2: `#9f71e3`

### Dark Mode
- Primary: `#0A0F1C`
- Secondary: `#EDEDED`
- Accent 1: `#00D9C0`
- Accent 2: `#A3364A`

## Getting Started

### Prerequisites
- Node.js 18+
- npm or yarn

### Installation

1. Navigate to the landing page directory:
```bash
cd src/landing-page
```

2. Install dependencies:
```bash
npm install
```

3. Start the development server:
```bash
npm run dev
```

4. Open your browser to `http://localhost:5173`

### Building for Production

```bash
npm run build
```

The built files will be in the `dist` directory.

## Project Structure

```
src/landing-page/
├── src/
│   ├── components/
│   │   ├── LandingPageSections.tsx    # Header, Hero, Features
│   │   └── LandingPageExtended.tsx    # HowItWorks, Community, Roadmap, Footer
│   ├── lib/
│   │   └── utils.ts                   # Utility functions
│   ├── App.tsx                        # Main app component
│   ├── main.tsx                       # React entry point
│   ├── index.css                      # Global styles and theme
│   └── vite-env.d.ts                  # Type declarations
├── public/
│   └── favicon.ico                    # Libr favicon
├── index.html                         # HTML template
├── package.json                       # Dependencies
├── tailwind.config.ts                 # Tailwind configuration
├── tsconfig.json                      # TypeScript config
└── vite.config.ts                     # Vite configuration
```

## Key Components

### Header
- Fixed navigation with smooth scroll
- Responsive hamburger menu
- Call-to-action button

### Hero Section
- Compelling headline with gradient text
- Feature highlights with animations
- Interactive demo preview
- Social proof indicators

### Features Section
- Grid of key features with icons
- Hover animations and effects
- Clear benefit descriptions

### How It Works
- Step-by-step process explanation
- Numbered progression
- Visual hierarchy

### Community Section
- User testimonials
- Community statistics
- Social links and engagement

### Roadmap
- Development phases with status
- Timeline visualization
- Feature checklists

### Footer
- Links organization
- Newsletter signup
- Social media links
- Copyright information

## Customization

### Colors
Edit the CSS custom properties in `src/index.css` to change the theme colors.

### Content
Update the content in the component files:
- `LandingPageSections.tsx` - Hero, Features
- `LandingPageExtended.tsx` - Community, Roadmap

### Animations
Framer Motion animations can be customized in each component. Common patterns include:
- Fade in on scroll
- Slide in from sides
- Scale and rotation effects
- Stagger animations for lists

## Performance

- **Lazy Loading**: Images and components load as needed
- **Code Splitting**: Automatic with Vite
- **Optimized Assets**: Compressed images and fonts
- **Minimal Bundle**: Tree shaking removes unused code

## Accessibility

- **Semantic HTML**: Proper heading hierarchy and landmarks
- **Keyboard Navigation**: All interactive elements accessible
- **Screen Reader Support**: ARIA labels and descriptions
- **Color Contrast**: Meets WCAG guidelines
- **Focus Management**: Visible focus indicators

## Contributing

1. Follow the existing code style and patterns
2. Test responsiveness on multiple devices
3. Ensure accessibility standards are met
4. Update documentation for any new features

## License

This project is part of the Libr ecosystem and follows the same MIT license.
