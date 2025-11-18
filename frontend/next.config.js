/**
 * @type {import('next').NextConfig}
 */
const nextConfig = {
    output: "export",
    distDir: "dist",
    compress: false,
    
    // Optimizaciones para aplicación de escritorio con Wails
    images: {
      unoptimized: true,
    },
    
    // Deshabilitar características no necesarias en aplicación de escritorio
    swcMinify: true,
    reactStrictMode: true,
    
    // Optimización del proceso de compilación
    poweredByHeader: false,
    generateEtags: false,
  };
  
  module.exports = nextConfig;
  