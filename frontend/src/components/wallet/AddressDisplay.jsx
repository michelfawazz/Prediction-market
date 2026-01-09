import React, { useState } from 'react';

const AddressDisplay = ({ address, chain }) => {
    const [copied, setCopied] = useState(false);

    const copyToClipboard = async () => {
        try {
            await navigator.clipboard.writeText(address);
            setCopied(true);
            setTimeout(() => setCopied(false), 2000);
        } catch (err) {
            console.error('Failed to copy:', err);
        }
    };

    // Generate a simple QR-like visual (placeholder - in production use qrcode.react)
    const generateQRPlaceholder = () => {
        return (
            <div className="bg-white p-4 rounded-lg inline-block">
                <div className="w-40 h-40 bg-gray-200 flex items-center justify-center text-gray-500 text-xs text-center">
                    <div>
                        <svg className="w-16 h-16 mx-auto mb-2" fill="currentColor" viewBox="0 0 24 24">
                            <path d="M3 3h6v6H3V3zm2 2v2h2V5H5zm8-2h6v6h-6V3zm2 2v2h2V5h-2zM3 13h6v6H3v-6zm2 2v2h2v-2H5zm13-2h3v2h-3v-2zm-5 0h2v2h-2v-2zm2 2h2v2h-2v-2zm-2 2h2v2h-2v-2zm2 2h2v2h-2v-2zm2-2h3v4h-3v-4z"/>
                        </svg>
                        QR Code
                    </div>
                </div>
            </div>
        );
    };

    return (
        <div className="bg-gray-700 rounded-lg p-6 text-center">
            {/* QR Code Placeholder */}
            <div className="mb-4">
                {generateQRPlaceholder()}
            </div>

            {/* Address Display */}
            <div className="mb-4">
                <p className="text-gray-400 text-xs mb-2">
                    Deposit Address ({chain})
                </p>
                <div className="bg-gray-800 rounded-lg p-3 break-all">
                    <code className="text-white text-sm font-mono">{address}</code>
                </div>
            </div>

            {/* Copy Button */}
            <button
                onClick={copyToClipboard}
                className={`px-6 py-2 rounded-lg font-medium transition-colors
                    ${copied
                        ? 'bg-green-600 text-white'
                        : 'bg-blue-600 hover:bg-blue-700 text-white'}`}
            >
                {copied ? 'Copied!' : 'Copy Address'}
            </button>
        </div>
    );
};

export default AddressDisplay;
