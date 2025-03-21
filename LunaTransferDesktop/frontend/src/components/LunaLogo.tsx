import React from 'react';
import { motion } from 'framer-motion';

interface LogoProps {
  size?: number;
  className?: string;
}

const LunaLogo: React.FC<LogoProps> = ({ size = 40, className = '' }) => {
  return (
    <div className={`relative ${className}`} style={{ width: size, height: size }}>
      <motion.div
        className="absolute inset-0 bg-indigo-600 rounded-full opacity-20"
        animate={{ scale: [1, 1.1, 1] }}
        transition={{ 
          repeat: Infinity, 
          duration: 2,
          ease: "easeInOut" 
        }}
      />
      <motion.div 
        className="absolute inset-0 flex items-center justify-center"
        initial={{ rotate: 0 }}
        animate={{ rotate: 360 }}
        transition={{ 
          repeat: Infinity, 
          duration: 30,
          ease: "linear" 
        }}
      >
        <div className="w-full h-full rounded-full border-t-2 border-r-2 border-indigo-200 opacity-50" />
      </motion.div>
      <div className="absolute inset-0 flex items-center justify-center">
        <motion.div 
          className="w-3/4 h-3/4 bg-indigo-500 rounded-full shadow-lg flex items-center justify-center text-white font-bold"
          whileHover={{ scale: 1.1 }}
        >
          LT
        </motion.div>
      </div>
    </div>
  );
};

export default LunaLogo;