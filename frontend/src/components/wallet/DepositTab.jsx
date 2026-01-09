import React, { useState, useEffect } from 'react';
import { API_URL } from '../../config';
import ChainSelector, { chains } from './ChainSelector';
import AddressDisplay from './AddressDisplay';

const DepositTab = () => {
    const [selectedChain, setSelectedChain] = useState(chains[0]?.id || 'ethereum');
    const [depositAddress, setDepositAddress] = useState(null);
    const [loading, setLoading] = useState(false);
    const [error, setError] = useState(null);

    useEffect(() => {
        fetchDepositAddress(selectedChain);
    }, [selectedChain]);

    const fetchDepositAddress = async (chain) => {
        setLoading(true);
        setError(null);
        try {
            const token = localStorage.getItem('token');
            const response = await fetch(`${API_URL}/v0/wallet/deposit/${chain}`, {
                headers: {
                    'Authorization': `Bearer ${token}`,
                },
            });
            if (!response.ok) {
                const errorText = await response.text();
                throw new Error(errorText || 'Failed to fetch deposit address');
            }
            const data = await response.json();
            setDepositAddress(data);
        } catch (err) {
            setError(err.message);
            setDepositAddress(null);
        } finally {
            setLoading(false);
        }
    };

    return (
        <div className="space-y-6">
            {/* Instructions */}
            <div className="bg-gray-700 rounded-lg p-4">
                <h3 className="text-white font-medium mb-2">How to Deposit</h3>
                <ol className="text-gray-300 text-sm space-y-1 list-decimal list-inside">
                    <li>Select your preferred network below</li>
                    <li>Send USDC or USDT to the displayed address</li>
                    <li>Credits will be added automatically after confirmation</li>
                </ol>
            </div>

            {/* Chain Selector */}
            <div>
                <label className="block text-gray-400 text-sm mb-2">Select Network</label>
                <ChainSelector
                    selectedChain={selectedChain}
                    onSelect={setSelectedChain}
                    disabled={loading}
                />
            </div>

            {/* Address Display */}
            {loading ? (
                <div className="flex justify-center py-8">
                    <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-white"></div>
                </div>
            ) : error ? (
                <div className="bg-red-900/30 border border-red-600/50 rounded-lg p-4 text-center">
                    <p className="text-red-400">{error}</p>
                    <button
                        onClick={() => fetchDepositAddress(selectedChain)}
                        className="mt-2 text-sm text-blue-400 hover:text-blue-300"
                    >
                        Try again
                    </button>
                </div>
            ) : depositAddress ? (
                <AddressDisplay
                    address={depositAddress.address}
                    chain={depositAddress.displayName || selectedChain}
                />
            ) : null}

            {/* Supported Tokens Info */}
            <div className="bg-gray-700 rounded-lg p-4">
                <h4 className="text-white text-sm font-medium mb-2">Supported Tokens</h4>
                <div className="flex space-x-4">
                    <div className="flex items-center space-x-2">
                        <div className="w-6 h-6 bg-blue-500 rounded-full flex items-center justify-center text-white text-xs font-bold">$</div>
                        <span className="text-gray-300 text-sm">USDC</span>
                    </div>
                    <div className="flex items-center space-x-2">
                        <div className="w-6 h-6 bg-green-500 rounded-full flex items-center justify-center text-white text-xs font-bold">$</div>
                        <span className="text-gray-300 text-sm">USDT</span>
                    </div>
                </div>
                <p className="text-gray-500 text-xs mt-2">1 USDC/USDT = 1 Credit</p>
            </div>

            {/* Warning */}
            <div className="bg-yellow-900/30 border border-yellow-600/50 rounded-lg p-3">
                <p className="text-yellow-400 text-xs">
                    Only send USDC or USDT on the selected network. Sending other tokens or using
                    the wrong network may result in permanent loss of funds.
                </p>
            </div>
        </div>
    );
};

export default DepositTab;
