import React from 'react';

// Set to true to show testnet options, false for mainnet only
const SHOW_TESTNETS = import.meta.env.VITE_SHOW_TESTNETS === 'true';

const mainnetChains = [
    { id: 'ethereum', name: 'Ethereum', chainId: 1, icon: 'ETH' },
    { id: 'tron', name: 'TRON', chainId: 728126428, icon: 'TRX' },
];

const testnetChains = [
    { id: 'ethereum-sepolia', name: 'Ethereum Sepolia', chainId: 11155111, icon: 'ETH', testnet: true },
    { id: 'tron-nile', name: 'TRON Nile', chainId: 3448148188, icon: 'TRX', testnet: true },
];

const chains = SHOW_TESTNETS ? testnetChains : mainnetChains;

const ChainSelector = ({ selectedChain, onSelect, disabled = false }) => {
    return (
        <div className="grid grid-cols-2 gap-2">
            {chains.map((chain) => (
                <button
                    key={chain.id}
                    type="button"
                    onClick={() => !disabled && onSelect(chain.id)}
                    disabled={disabled}
                    className={`flex flex-col items-center justify-center p-3 rounded-lg border transition-colors
                        ${disabled ? 'opacity-50 cursor-not-allowed' : ''}
                        ${selectedChain === chain.id
                            ? 'bg-blue-600/20 border-blue-500 text-white'
                            : 'bg-gray-700 border-gray-600 text-gray-300 hover:border-gray-500'
                        }`}
                >
                    <span className="text-lg font-bold">{chain.icon}</span>
                    <span className="text-sm font-medium">{chain.name}</span>
                    {chain.testnet && (
                        <span className="text-xs text-yellow-400 mt-1">Testnet</span>
                    )}
                </button>
            ))}
        </div>
    );
};

// Export chains for use in other components
export { chains, mainnetChains, testnetChains };
export default ChainSelector;
