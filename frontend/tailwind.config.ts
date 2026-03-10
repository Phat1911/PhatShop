import type { Config } from "tailwindcss";

const config: Config = {
  content: [
    "./pages/**/*.{js,ts,jsx,tsx,mdx}",
    "./components/**/*.{js,ts,jsx,tsx,mdx}",
    "./app/**/*.{js,ts,jsx,tsx,mdx}",
  ],
  theme: {
    extend: {
      colors: {
        background: "#0a0a0f",
        "bg-secondary": "#111118",
        "bg-card": "#16161e",
        "accent-red": "#e63946",
        "accent-blue": "#2563eb",
        "border-dark": "#1f1f2e",
        "text-muted": "#9ca3af",
      },
      animation: {
        "fade-in": "fadeIn 0.4s ease both",
        "slide-down": "slideDown 0.25s ease both",
      },
    },
  },
  plugins: [],
};
export default config;
