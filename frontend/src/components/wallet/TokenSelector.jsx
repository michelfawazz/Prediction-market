import React from 'react';

const tokens = [
    { symbol: 'USDC', name: 'USD Coin' },
    { symbol: 'USDT', name: 'Tether USD' },
];

const TokenSelector = ({ selectedToken, onSelect, disabled = false }) => {
    return (
        <div className="flex space-x-2">
            {tokens.map((token) => (
                <button
                    key={token.symbol}
                    type="button"
                    onClick={() => !disabled && onSelect(token.symbol)}
                    disabled={disabled}
                    className={`flex-1 flex items-center justify-center space-x-2 p-3 rounded-lg border transition-colors
                        ${disabled ? 'opacity-50 cursor-not-allowed' : ''}
                        ${selectedToken === token.symbol
                            ? 'bg-blue-600/20 border-blue-500 text-white'
                            : 'bg-gray-700 border-gray-600 text-gray-300 hover:border-gray-500'
                        }`}
                >
                    <span className="text-sm font-medium">{token.symbol}</span>
                </button>
            ))}
        </div>
    );
};

export default TokenSelector;
