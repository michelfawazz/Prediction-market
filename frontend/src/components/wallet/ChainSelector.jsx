import React from 'react';

const chains = [
    { id: 'ethereum', name: 'Ethereum', chainId: 1 },
    { id: 'polygon', name: 'Polygon', chainId: 137 },
    { id: 'base', name: 'Base', chainId: 8453 },
    { id: 'arbitrum', name: 'Arbitrum', chainId: 42161 },
];

const ChainSelector = ({ selectedChain, onSelect, disabled = false }) => {
    return (
        <div className="grid grid-cols-2 gap-2">
            {chains.map((chain) => (
                <button
                    key={chain.id}
                    type="button"
                    onClick={() => !disabled && onSelect(chain.id)}
                    disabled={disabled}
                    className={`flex items-center justify-center space-x-2 p-3 rounded-lg border transition-colors
                        ${disabled ? 'opacity-50 cursor-not-allowed' : ''}
                        ${selectedChain === chain.id
                            ? 'bg-blue-600/20 border-blue-500 text-white'
                            : 'bg-gray-700 border-gray-600 text-gray-300 hover:border-gray-500'
                        }`}
                >
                    <span className="text-sm font-medium">{chain.name}</span>
                </button>
            ))}
        </div>
    );
};

export default ChainSelector;
