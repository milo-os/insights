import type { Config } from "tailwindcss"

const config = {
  darkMode: ["class"],
  content: [
    './src/pages/**/*.{ts,tsx}',
    './src/components/**/*.{ts,tsx}',
    './src/app/**/*.{ts,tsx}',
  ],
  prefix: "",
  theme: {
    container: {
      center: true,
      padding: "2rem",
      screens: {
        "2xl": "1400px",
      },
    },
    extend: {
      colors: {
        // Datum Brand Colors
        "midnight-fjord": {
          DEFAULT: "#0C1D31",
          900: "#0C1D31",
          800: "#152A42",
          700: "#1E3753",
          600: "#274464",
          500: "#305175",
        },
        "aurora-moss": {
          DEFAULT: "#E6F59F",
          900: "#E6F59F",
          800: "#EAF7B0",
          700: "#EEF9C1",
          600: "#F2FBD2",
          500: "#F6FDE3",
        },
        "blush-quartz": {
          DEFAULT: "#ECD0D0",
          900: "#ECD0D0",
          800: "#F0DADA",
          700: "#F4E4E4",
          600: "#F8EEEE",
          500: "#FCF8F8",
        },
        "canyon-clay": {
          DEFAULT: "#BF9595",
          900: "#BF9595",
          800: "#CCAAAA",
          700: "#D9BFBF",
          600: "#E6D4D4",
          500: "#F3E9E9",
        },
        "pine-forge": {
          DEFAULT: "#4D6356",
          900: "#4D6356",
          800: "#617768",
          700: "#758B7A",
          600: "#899F8C",
          500: "#9DB39E",
        },
        "glacier-mist": {
          DEFAULT: "#E8E7E4",
          900: "#E8E7E4",
          800: "#EFEFED",
          700: "#F6F6F5",
          600: "#FAFAFA",
          500: "#FFFFFF",
        },
        // Semantic colors mapped to brand
        border: "hsl(var(--border))",
        input: "hsl(var(--input))",
        ring: "hsl(var(--ring))",
        background: "hsl(var(--background))",
        foreground: "hsl(var(--foreground))",
        primary: {
          DEFAULT: "hsl(var(--primary))",
          foreground: "hsl(var(--primary-foreground))",
        },
        secondary: {
          DEFAULT: "hsl(var(--secondary))",
          foreground: "hsl(var(--secondary-foreground))",
        },
        destructive: {
          DEFAULT: "hsl(var(--destructive))",
          foreground: "hsl(var(--destructive-foreground))",
        },
        muted: {
          DEFAULT: "hsl(var(--muted))",
          foreground: "hsl(var(--muted-foreground))",
        },
        accent: {
          DEFAULT: "hsl(var(--accent))",
          foreground: "hsl(var(--accent-foreground))",
        },
        popover: {
          DEFAULT: "hsl(var(--popover))",
          foreground: "hsl(var(--popover-foreground))",
        },
        card: {
          DEFAULT: "hsl(var(--card))",
          foreground: "hsl(var(--card-foreground))",
        },
        // Severity colors
        severity: {
          info: "#3B82F6",
          warning: "#F59E0B",
          critical: "#EF4444",
        },
      },
      borderRadius: {
        lg: "var(--radius)",
        md: "calc(var(--radius) - 2px)",
        sm: "calc(var(--radius) - 4px)",
      },
      fontFamily: {
        sans: ["var(--font-alliance)", "system-ui", "sans-serif"],
        display: ["var(--font-canela)", "Georgia", "serif"],
      },
      keyframes: {
        "accordion-down": {
          from: { height: "0" },
          to: { height: "var(--radix-accordion-content-height)" },
        },
        "accordion-up": {
          from: { height: "var(--radix-accordion-content-height)" },
          to: { height: "0" },
        },
      },
      animation: {
        "accordion-down": "accordion-down 0.2s ease-out",
        "accordion-up": "accordion-up 0.2s ease-out",
      },
    },
  },
  plugins: [require("tailwindcss-animate")],
} satisfies Config

export default config
