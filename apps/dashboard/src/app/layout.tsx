import type { Metadata } from 'next';
import './globals.css';

export const metadata: Metadata = {
  title: 'Antisky — Cloud Hosting Dashboard',
  description: 'Deploy, manage, and scale your applications with Antisky. Multi-language hosting for Node.js, Go, Python, PHP, and more.',
  icons: { icon: '/favicon.ico' },
};

export default function RootLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  return (
    <html lang="en">
      <head>
        <script src="https://checkout.razorpay.com/v1/checkout.js" async></script>
      </head>
      <body>{children}</body>
    </html>
  );
}
