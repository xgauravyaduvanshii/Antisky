import type { Metadata } from 'next';
import './globals.css';

export const metadata: Metadata = {
  title: 'Antisky Admin — Platform Control',
  description: 'Full-control admin panel for the Antisky hosting platform. Manage servers, users, deployments, and infrastructure.',
};

export default function RootLayout({ children }: { children: React.ReactNode }) {
  return (
    <html lang="en">
      <head>
        <script src="https://checkout.razorpay.com/v1/checkout.js" async></script>
      </head>
      <body>{children}</body>
    </html>
  );
}
