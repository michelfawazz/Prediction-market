import React, { useState } from 'react';
import ReactDOM from 'react-dom';
import DepositTab from './DepositTab';
import WithdrawTab from './WithdrawTab';
import TransactionHistory from './TransactionHistory';

const WalletModal = ({ isOpen, onClose }) => {
    const [activeTab, setActiveTab] = useState('deposit');

    if (!isOpen) return null;

    const modalContent = (
        <div className="fixed inset-0 bg-black bg-opacity-50 flex justify-center items-center z-50">
            <div className="relative bg-gray-800 rounded-lg w-full max-w-lg mx-4 max-h-[90vh] overflow-hidden">
                {/* Header */}
                <div className="flex items-center justify-between p-4 border-b border-gray-700">
                    <h2 className="text-xl font-bold text-white">Wallet</h2>
                    <button
                        onClick={onClose}
                        className="text-gray-400 hover:text-white text-2xl leading-none"
                    >
                        &times;
                    </button>
                </div>

                {/* Tabs */}
                <div className="flex border-b border-gray-700">
                    <button
                        onClick={() => setActiveTab('deposit')}
                        className={`flex-1 py-3 px-4 text-sm font-medium transition-colors
                            ${activeTab === 'deposit'
                                ? 'text-green-400 border-b-2 border-green-400'
                                : 'text-gray-400 hover:text-white'}`}
                    >
                        Deposit
                    </button>
                    <button
                        onClick={() => setActiveTab('withdraw')}
                        className={`flex-1 py-3 px-4 text-sm font-medium transition-colors
                            ${activeTab === 'withdraw'
                                ? 'text-red-400 border-b-2 border-red-400'
                                : 'text-gray-400 hover:text-white'}`}
                    >
                        Withdraw
                    </button>
                    <button
                        onClick={() => setActiveTab('history')}
                        className={`flex-1 py-3 px-4 text-sm font-medium transition-colors
                            ${activeTab === 'history'
                                ? 'text-blue-400 border-b-2 border-blue-400'
                                : 'text-gray-400 hover:text-white'}`}
                    >
                        History
                    </button>
                </div>

                {/* Content */}
                <div className="p-4 overflow-y-auto max-h-[60vh]">
                    {activeTab === 'deposit' && <DepositTab />}
                    {activeTab === 'withdraw' && <WithdrawTab onClose={onClose} />}
                    {activeTab === 'history' && <TransactionHistory />}
                </div>
            </div>
        </div>
    );

    // Use portal to render modal at document body level
    const modalRoot = document.getElementById('modal-root');
    if (modalRoot) {
        return ReactDOM.createPortal(modalContent, modalRoot);
    }

    // Fallback if modal-root doesn't exist
    return modalContent;
};

export default WalletModal;
