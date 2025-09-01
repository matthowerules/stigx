import React, { useState, useEffect } from 'react';
import './App.css';
import { OpenFileDialog, AnalyzeFile, ValidateFile, ConvertFile, SaveFileDialog, ImportStigBundleDialog, ImportStigBundle, GetImportedStigs, DeleteImportedStigs, CreateChecklist, CreateChecklistFromStigs, LoadChecklistFromFile, SaveChecklistToFile } from "../wailsjs/go/main/App";
import { main, models, services } from "../wailsjs/go/models";
import { EventsOn } from "../wailsjs/runtime/runtime";

interface ImportProgress {
  stage: string;
  currentFile: string;
  processedFiles: number;
  totalFiles: number;
  processedStigs: number;
  totalStigs: number;
  message: string;
  error?: string;
}

interface ChecklistTab {
  id: string;
  name: string;
  checklist: any;
  items: any[];
  selectedRule: any;
}

interface TabBarProps {
  tabs: ChecklistTab[];
  activeTabId: string | null;
  onTabSelect: (tabId: string) => void;
  onTabClose: (tabId: string) => void;
}

function TabBar({ tabs, activeTabId, onTabSelect, onTabClose }: TabBarProps) {
  return (
    <div className="tab-bar">
      {tabs.map(tab => (
        <div
          key={tab.id}
          className={`tab ${activeTabId === tab.id ? 'active' : ''}`}
          onClick={() => onTabSelect(tab.id)}
        >
          <span className="tab-title">{tab.name}</span>
          <button
            className="tab-close"
            onClick={(e) => {
              e.stopPropagation();
              onTabClose(tab.id);
            }}
          >
            ×
          </button>
        </div>
      ))}
    </div>
  );
}

function App() {
  const [selectedFile, setSelectedFile] = useState<string>('');
  const [fileInfo, setFileInfo] = useState<main.FileInfo | null>(null);
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState<string>('');
  const [showStigList, setShowStigList] = useState(false);
  const [stigList, setStigList] = useState<any[]>([]);
  const [selectedStigs, setSelectedStigs] = useState<Set<number>>(new Set());
  const [lastSelectedIndex, setLastSelectedIndex] = useState<number | null>(null);
  const [showCreateChecklist, setShowCreateChecklist] = useState(false);
  const [showChecklistViewer, setShowChecklistViewer] = useState(false);
  const [checklistTabs, setChecklistTabs] = useState<ChecklistTab[]>([]);
  const [activeTabId, setActiveTabId] = useState<string | null>(null);
  const [showImportProgress, setShowImportProgress] = useState(false);
  const [importProgress, setImportProgress] = useState<ImportProgress | null>(null);

  useEffect(() => {
    const unsubscribe = EventsOn("import-progress", (progress: ImportProgress) => {
      setImportProgress(progress);
      setShowImportProgress(true);

      if (progress.stage === "complete") {
        setTimeout(() => {
          setShowImportProgress(false);
          setImportProgress(null);
        }, 2000); // Show completion message for 2 seconds
      }
    });

    return unsubscribe;
  }, []);

  const handleOpenFile = async () => {
    try {
      setIsLoading(true);
      setError('');
      const filePath = await OpenFileDialog();
      if (filePath) {
        setSelectedFile(filePath);
        const info = await AnalyzeFile(filePath);
        setFileInfo(info);
      }
    } catch (err) {
      setError(`Failed to open file: ${err}`);
    } finally {
      setIsLoading(false);
    }
  };

  const handleLoadChecklist = async () => {
    try {
      setIsLoading(true);
      setError('');

      const filePath = await OpenFileDialog();
      if (filePath) {
        const checklistData = await LoadChecklistFromFile(filePath);
        if (checklistData && checklistData.Checklist) {
          const fileName = filePath.split(/[\\/]/).pop() || 'Unnamed Checklist';
          const newTab: ChecklistTab = {
            id: Date.now().toString(),
            name: fileName,
            checklist: checklistData.Checklist,
            items: checklistData.Items || [],
            selectedRule: checklistData.Items && checklistData.Items.length > 0 ? checklistData.Items[0] : null
          };

          setChecklistTabs(prev => [...prev, newTab]);
          setActiveTabId(newTab.id);
          setShowChecklistViewer(true);
          setShowStigList(false); // Hide STIG list when loading checklist

          alert(`Successfully loaded checklist: ${checklistData.Checklist.name}`);
        } else {
          setError('Failed to load checklist: Invalid file format');
        }
      }
    } catch (err) {
      setError(`Failed to load checklist: ${err}`);
    } finally {
      setIsLoading(false);
    }
  };

  const handleValidateFile = async () => {
    if (!selectedFile) return;

    try {
      setIsLoading(true);
      setError('');
      const info = await ValidateFile(selectedFile);
      setFileInfo(info);
    } catch (err) {
      setError(`Failed to validate file: ${err}`);
    } finally {
      setIsLoading(false);
    }
  };

  const handleConvertFile = async () => {
    if (!selectedFile || !fileInfo) return;

    try {
      setIsLoading(true);
      setError('');

      // Determine output extension based on current format
      const outputExt = fileInfo.format === 'CKL' ? '.cklb' : '.ckl';
      const baseName = selectedFile.split('/').pop()?.replace(/\.[^/.]+$/, "") || "converted";
      const defaultFilename = `${baseName}_converted${outputExt}`;

      const outputPath = await SaveFileDialog(defaultFilename);
      if (!outputPath) {
        setError('No output file selected.');
        return;
      }

      // Conversion logic
      const result = await ConvertFile(selectedFile, outputPath);
      if (result.success) {
        alert(`File converted successfully!\nOutput: ${result.outputPath}\nSize reduction: ${Math.round((1 - result.outputSize / result.inputSize) * 100)}%`);
      } else {
        setError(`Conversion failed: ${result.errorMessage}`);
      }
    } catch (err) {
      setError(`Failed to convert file: ${err}`);
    } finally {
      setIsLoading(false);
    }
  };

  const handleImportStigs = async () => {
    try {
      setIsLoading(true);
      setError('');

      const filePath = await ImportStigBundleDialog();
      if (filePath) {
        setShowImportProgress(true);
        setImportProgress({
          stage: "starting",
          message: `Starting import of ${filePath.split('/').pop()}...`,
          currentFile: "",
          processedFiles: 0,
          totalFiles: 0,
          processedStigs: 0,
          totalStigs: 0
        });

        await ImportStigBundle(filePath);

        if (showStigList) {
          await loadStigList();
        }

      }
    } catch (err) {
      setError(`Failed to import STIGs: ${err}`);
      setShowImportProgress(false);
      setImportProgress(null);
    } finally {
      setIsLoading(false);
    }
  };

  const loadStigList = async () => {
    try {
      const stigs = await GetImportedStigs();
      setStigList(stigs || []);
    } catch (err) {
      setError(`Failed to load STIGs: ${err}`);
      setStigList([]);
    }
  };

  const handleViewStigs = async () => {
    try {
      setError('');
      setShowStigList(!showStigList);

      if (!showStigList) {
        await loadStigList();
      }
    } catch (err) {
      setError(`Failed to load STIGs: ${err}`);
    }
  };

  const handleStigSelection = (stigId: number, checked: boolean, index?: number, shiftKey?: boolean) => {
    const newSelected = new Set(selectedStigs);

    if (shiftKey && lastSelectedIndex !== null && index !== undefined) {
      // Handle shift-click range selection then select all in the range
      const startIndex = Math.min(lastSelectedIndex, index);
      const endIndex = Math.max(lastSelectedIndex, index);

      for (let i = startIndex; i <= endIndex; i++) {
        const stig = stigList[i];
        if (stig) {
          newSelected.add(stig.id);
        }
      }
    } else {
      // Handle regular selection
      if (checked) {
        newSelected.add(stigId);
      } else {
        newSelected.delete(stigId);
      }
      setLastSelectedIndex(index ?? null);
    }

    setSelectedStigs(newSelected);
  };

  const handleStigItemClick = (stigId: number, index: number, event: React.MouseEvent) => {
    event.preventDefault();
    const isSelected = selectedStigs.has(stigId);
    handleStigSelection(stigId, !isSelected, index, event.shiftKey);
  };

  const handleCreateChecklist = () => {
    if (selectedStigs.size === 0) {
      setError('Please select at least one STIG to create a checklist.');
      return;
    }
    setShowCreateChecklist(true);
  };

  const handleSelectAll = () => {
    if (selectedStigs.size === stigList.length) {
      setSelectedStigs(new Set());
      setLastSelectedIndex(null);
    } else {
      const allIds = new Set(stigList.map(stig => stig.id));
      setSelectedStigs(allIds);
      setLastSelectedIndex(stigList.length - 1);
    }
  };

  const handleDeleteStigs = async () => {
    if (selectedStigs.size === 0) {
      setError('Please select at least one STIG to delete.');
      return;
    }

    const selectedCount = selectedStigs.size;
    const selectedNames = Array.from(selectedStigs).map(id =>
      stigList.find(s => s.id === id)?.name || `ID: ${id}`
    ).join(', ');

    const confirmed = window.confirm(
      `Are you sure you want to delete ${selectedCount} STIG${selectedCount !== 1 ? 's' : ''}?\n\n` +
      `Selected STIGs: ${selectedNames}\n\n` +
      `This action cannot be undone.`
    );

    if (!confirmed) {
      return;
    }

    try {
      setIsLoading(true);
      setError('');

      const stigIdsToDelete = Array.from(selectedStigs);
      await DeleteImportedStigs(stigIdsToDelete);

      setSelectedStigs(new Set());
      setLastSelectedIndex(null);

      await loadStigList();

      alert(`Successfully deleted ${selectedCount} STIG${selectedCount !== 1 ? 's' : ''}.`);

    } catch (err) {
      setError(`Failed to delete STIGs: ${err}`);
    } finally {
      setIsLoading(false);
    }
  };

  const handleChecklistSubmit = async (checklistData: any) => {
    try {
      setIsLoading(true);
      setError('');

      // Prompt user for save location first to get the checklist name - need to set the extension in the dialog properly
      const defaultName = checklistData.name ? `${checklistData.name}.cklb` : 'checklist.cklb';
      const outputPath = await SaveFileDialog(defaultName);

      if (!outputPath) {
        setShowCreateChecklist(false);
        setSelectedStigs(new Set());
        return;
      }

      const fileName = outputPath.split(/[\\/]/).pop() || 'checklist.cklb';
      const checklistName = fileName.replace(/\.[^/.]+$/, ''); // Remove file extension

      // Create full checklist data structure from selected STIGs
      const request = services.CreateChecklistRequest.createFrom({
        name: checklistName, // Use filename as checklist name
        description: checklistData.description,
        stigIds: Array.from(selectedStigs),
        targetData: checklistData.targetData,
        format: 'cklb'
      });

      const fullChecklistData = await CreateChecklistFromStigs(request);
      await SaveChecklistToFile(fullChecklistData, 'cklb', outputPath);
      const result = await CreateChecklist(request);

      // Load the created checklist and add as a new tab
      try {
        const checklistData = await LoadChecklistFromFile(outputPath);
        if (checklistData && checklistData.Checklist) {
          const fileName = outputPath.split(/[\\/]/).pop() || 'Unnamed Checklist';
          const newTab: ChecklistTab = {
            id: Date.now().toString(),
            name: fileName,
            checklist: checklistData.Checklist,
            items: checklistData.Items || [],
            selectedRule: checklistData.Items && checklistData.Items.length > 0 ? checklistData.Items[0] : null
          };

          setChecklistTabs(prev => [...prev, newTab]);
          setActiveTabId(newTab.id);
          setShowChecklistViewer(true);
          setShowStigList(false); // Hide STIG list when creating checklist
          setShowCreateChecklist(false);
          setSelectedStigs(new Set());

          alert(
            `Checklist "${result.name}" created and saved successfully!\n` +
            `${result.stigCount} STIGs, ${result.ruleCount} total rules\n` +
            `Saved to: ${outputPath}`
          );
        } else {
          throw new Error('Failed to load created checklist');
        }
      } catch (loadErr) {
        // If loading fails, still show success but don't auto-open viewer
        alert(
          `Checklist "${result.name}" created and saved successfully!\n` +
          `${result.stigCount} STIGs, ${result.ruleCount} total rules\n` +
          `Saved to: ${outputPath}\n\n` +
          `Note: Could not auto-open viewer. You can manually load the file using "Load Checklist".`
        );
        setShowCreateChecklist(false);
        setSelectedStigs(new Set());
      }

    } catch (err) {
      setError(`Failed to create checklist: ${err}`);
    } finally {
      setIsLoading(false);
    }
  };

  // Keyboard shortcuts for STIG selection
  useEffect(() => {
    const handleKeyDown = (event: KeyboardEvent) => {
      if (!showStigList || stigList.length === 0) return;

      if (event.ctrlKey || event.metaKey) {
        switch (event.key) {
          case 'a':
            event.preventDefault();
            handleSelectAll();
            break;
          case 'd':
            event.preventDefault();
            if (selectedStigs.size > 0) {
              handleDeleteStigs();
            }
            break;
        }
      } else {
        switch (event.key) {
          case 'Escape':
            if (selectedStigs.size > 0) {
              setSelectedStigs(new Set());
              setLastSelectedIndex(null);
            }
            break;
          case 'Delete':
            if (selectedStigs.size > 0) {
              handleDeleteStigs();
            }
            break;
        }
      }
    };

    document.addEventListener('keydown', handleKeyDown);
    return () => document.removeEventListener('keydown', handleKeyDown);
  }, [showStigList, stigList, selectedStigs, handleSelectAll, handleDeleteStigs]);

  const handleViewChecklists = async () => {
    try {
      setError('');
      setShowChecklistViewer(!showChecklistViewer);

      if (!showChecklistViewer) {
        // If no tabs are loaded, prompt to load one
        if (checklistTabs.length === 0) {
          await handleLoadChecklist();
        }
      } else {
        // Reset checklist viewer state when hiding
        setChecklistTabs([]);
        setActiveTabId(null);
      }
    } catch (err) {
      setError(`Failed to load checklists: ${err}`);
    } finally {
      setIsLoading(false);
    }
  };

  const handleRuleSelection = (rule: any) => {
    setChecklistTabs(prev => prev.map(tab =>
      tab.id === activeTabId ? { ...tab, selectedRule: rule } : tab
    ));
  };

  const handleUpdateRuleStatus = async (status: string, findingDetails: string, comments: string, ruleId?: string) => {
    const activeTab = checklistTabs.find(tab => tab.id === activeTabId);
    if (!activeTab) return;

    // Use provided ruleId for auto-save or create a unique identifier for the current rule
    const targetRuleId = ruleId || `${activeTab.selectedRule?.Item.vulnId}-${activeTab.selectedRule?.STIG.id}`;
    if (!targetRuleId) return;

    try {
      setError('');

      // Update local state only (since we're working with files, changes are in-memory)
      setChecklistTabs(prev => prev.map(tab => {
        if (tab.id === activeTabId) {
          const updatedItems = tab.items.map(item => {
            const itemId = `${item.Item.vulnId}-${item.STIG.id}`;
            return itemId === targetRuleId
              ? { ...item, Item: { ...item.Item, status, findingDetails, comments } }
              : item;
          });

          // Update selected rule only if we're updating the currently selected rule
          let updatedSelectedRule = tab.selectedRule;
          if (!ruleId && tab.selectedRule) {
            const selectedRuleId = `${tab.selectedRule.Item.vulnId}-${tab.selectedRule.STIG.id}`;
            if (selectedRuleId === targetRuleId) {
              updatedSelectedRule = {
                ...tab.selectedRule,
                Item: { ...tab.selectedRule.Item, status, findingDetails, comments }
              };
            }
          }

          return {
            ...tab,
            items: updatedItems,
            selectedRule: updatedSelectedRule
          };
        }
        return tab;
      }));

    } catch (err) {
      setError(`Failed to update rule status: ${err}`);
    }
  };

  const formatFileSize = (bytes: number): string => {
    if (bytes === 0) return '0 Bytes';
    const k = 1024;
    const sizes = ['Bytes', 'KB', 'MB', 'GB'];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
  };

  return (
    <div className="app">
      {/* I go back and forth on whether to keep this header or not - I like it as an intro, but not for workflows after.  Maybe at the bottom????
            <header className="app-header">
                <h1>STIG Viewer X</h1>
                <p>Modern cross-platform STIG/CKL viewer and editor</p>
            </header>
            */}

      <div className="file-operations">
        <button
          onClick={handleImportStigs} // Need to add the action to hide the header element - nvm, gonna leave it, was a bad move.
          className="btn btn-primary"
        >
          Import STIGs
        </button>
        <button
          onClick={handleViewStigs} // Need to add action to hide the header element when clicked to the function
          disabled={isLoading}
          className="btn btn-secondary"
        >
          {showStigList ? 'Hide STIGs' : 'View STIGs'}
        </button>
        {/* Commented out as redundant with Load Checklist button - may repurpose later
                <button
                    onClick={handleViewChecklists}
                    disabled={isLoading}
                    className="btn btn-secondary"
                >
                    {showChecklistViewer ? 'Hide Checklists' : 'View Checklists'}
                </button>
                */}
        <button
          onClick={handleLoadChecklist}
          disabled={isLoading}
          className="btn btn-primary"
        >
          {isLoading ? 'Loading...' : 'Load Checklist'}
        </button>

        {/* Past relic
                <button
                    onClick={handleOpenFile}
                    disabled={isLoading}
                    className="btn btn-secondary"
                >
                    {isLoading ? 'Loading...' : 'Validate CKL'}
                </button>
                */}

        {selectedFile && (
          <div className="file-actions">
            <button
              onClick={handleValidateFile}
              disabled={isLoading}
              className="btn btn-secondary"
            >
              Validate File
            </button>
            <button
              onClick={handleConvertFile}
              disabled={isLoading || !fileInfo?.isValid}
              className="btn btn-secondary"
            >
              Convert File
            </button>
          </div>
        )}
      </div>

      <main className="app-main">

        {showStigList && (
          <div className="stig-list">
            <h3>Imported STIGs</h3>
            {stigList.length === 0 ? (
              <p>No STIGs imported yet. Click "Import STIGs" to get started.</p>
            ) : (
              <>
                <div className="stig-list-container">
                  {stigList.map((stig, index) => (
                    <div
                      key={stig.id}
                      className={`stig-item ${selectedStigs.has(stig.id) ? 'selected' : ''}`}
                      onClick={(e) => handleStigItemClick(stig.id, index, e)}
                    >
                      <div className="stig-checkbox">
                        <input
                          type="checkbox"
                          id={`stig-${stig.id}`}
                          checked={selectedStigs.has(stig.id)}
                          onChange={(e) => {
                            e.stopPropagation();
                            handleStigSelection(stig.id, e.target.checked, index);
                          }}
                          onClick={(e) => e.stopPropagation()}
                        />
                      </div>
                      <div className="stig-content">
                        <div className="stig-name">{stig.name}</div>
                        <div className="stig-details">
                          <span className="stig-version">v{stig.version}</span>
                          <span className="stig-date">{stig.releaseDate || 'Unknown'}</span>
                          <span className="stig-rules">{stig.ruleCount} rules</span>
                        </div>
                      </div>
                    </div>
                  ))}
                </div>
                <div className="stig-help">
                  <small>
                    Click to select • Shift+click for range • Ctrl+A to select all • Esc to clear • Del to delete
                  </small>
                </div>
              </>
            )}
            {stigList.length > 0 && (
              <div className="stig-actions">
                <div className="stig-header">
                  <p className="selection-count">
                    {selectedStigs.size} of {stigList.length} STIG{stigList.length !== 1 ? 's' : ''} selected
                  </p>
                  <button
                    className="btn btn-secondary btn-small"
                    onClick={handleSelectAll}
                    disabled={stigList.length === 0}
                  >
                    {selectedStigs.size === stigList.length ? 'Clear All' : 'Select All'}
                  </button>
                </div>
                <div className="action-buttons">
                  <button
                    className="btn btn-success"
                    onClick={handleCreateChecklist}
                    disabled={selectedStigs.size === 0}
                  >
                    Create Checklist from Selected
                  </button>
                  <button
                    className="btn btn-danger"
                    onClick={handleDeleteStigs}
                    disabled={selectedStigs.size === 0 || isLoading}
                  >
                    Delete Selected STIGs
                  </button>
                </div>
              </div>
            )}
          </div>
        )}

        {showChecklistViewer && (
          <div className="checklist-viewer">
            {checklistTabs.length === 0 ? (
              <div className="checklist-list">
                <h3>Load Checklist</h3>
                <p>No checklist loaded. Click "Load Checklist" to open a CKL or CKLb file.</p>
              </div>
            ) : (
              <>
                <TabBar
                  tabs={checklistTabs}
                  activeTabId={activeTabId}
                  onTabSelect={setActiveTabId}
                  onTabClose={(tabId) => {
                    setChecklistTabs(prev => {
                      const newTabs = prev.filter(tab => tab.id !== tabId);
                      if (activeTabId === tabId && newTabs.length > 0) {
                        setActiveTabId(newTabs[0].id);
                      } else if (newTabs.length === 0) {
                        setActiveTabId(null);
                        setShowChecklistViewer(false);
                      }
                      return newTabs;
                    });
                  }}
                />
                {(() => {
                  const activeTab = checklistTabs.find(tab => tab.id === activeTabId);
                  return activeTab ? (
                    <ChecklistViewer
                      checklist={activeTab.checklist}
                      items={activeTab.items}
                      selectedRule={activeTab.selectedRule}
                      onRuleSelect={handleRuleSelection}
                      onUpdateRule={handleUpdateRuleStatus}
                      onBack={() => {
                        setChecklistTabs([]);
                        setActiveTabId(null);
                        setShowChecklistViewer(false);
                      }}
                    />
                  ) : null;
                })()}
              </>
            )}
          </div>
        )}

        {error && (
          <div className="error-message">
            <h3>Error</h3>
            <p>{error}</p>
          </div>
        )}

        {showCreateChecklist && (
          <CreateChecklistDialog
            selectedStigs={Array.from(selectedStigs).map(id =>
              stigList.find(s => s.id === id)
            ).filter(Boolean)}
            onSubmit={handleChecklistSubmit}
            onCancel={() => setShowCreateChecklist(false)}
          />
        )}

        {showImportProgress && importProgress && (
          <ProgressPopup
            progress={importProgress}
            onClose={() => {
              setShowImportProgress(false);
              setImportProgress(null);
            }}
          />
        )}

        {fileInfo && (
          <div className="file-info">
            <h2>File Information</h2>
            <div className="info-grid">
              <div className="info-item">
                <label>Name:</label>
                <span>{fileInfo.name}</span>
              </div>
              <div className="info-item">
                <label>Path:</label>
                <span className="path">{fileInfo.path}</span>
              </div>
              <div className="info-item">
                <label>Format:</label>
                <span className={`format format-${fileInfo.format.toLowerCase()}`}>
                  {fileInfo.format}
                </span>
              </div>
              <div className="info-item">
                <label>Size:</label>
                <span>{formatFileSize(fileInfo.size)}</span>
              </div>
              <div className="info-item">
                <label>Valid:</label>
                <span className={`status ${fileInfo.isValid ? 'valid' : 'invalid'}`}>
                  {fileInfo.isValid ? '✓ Valid' : '✗ Invalid'}
                </span>
              </div>
              {fileInfo.validationError && (
                <div className="info-item error">
                  <label>Error:</label>
                  <span>{fileInfo.validationError}</span>
                </div>
              )}
            </div>

            {fileInfo.statistics && (
              <div className="statistics">
                <h3>Statistics</h3>
                <div className="stats-grid">
                  {Object.entries(fileInfo.statistics).map(([key, value]) => (
                    <div key={key} className="stat-item">
                      <label>{key.replace(/_/g, ' ').replace(/\b\w/g, l => l.toUpperCase())}:</label>
                      <span>{String(value)}</span>
                    </div>
                  ))}
                </div>
              </div>
            )}
          </div>
        )}

        {!fileInfo && !error && !showStigList && !showChecklistViewer && (
          /*
          <div className="welcome">
              <h2>Welcome to STIG Viewer X</h2>
              <p>Click "Open CKL File" to get started with analyzing your existing CKL or CKLb files.</p>
              <ul>
                  <li>✓ CKL/CKLb file validation</li>
                  <li>✓ Format conversion (CKL ⇄ CKLb)</li>
                  <li>✓ STIG rule viewing and editing</li>
                  <li>✓ Cross-platform support</li>
              </ul>
          </div>
          */
          // I kinda like this here instead
          <header className="app-header">
            <h1>STIG Viewer X</h1>
            <p>Modern cross-platform STIG/CKL viewer and editor</p>
          </header>

        )}
      </main>
    </div>
  );
}

// CreateChecklistDialog component
interface CreateChecklistDialogProps {
  selectedStigs: any[];
  onSubmit: (data: any) => void;
  onCancel: () => void;
}

function CreateChecklistDialog({ selectedStigs, onSubmit, onCancel }: CreateChecklistDialogProps) {
  const [name, setName] = useState('');
  const [description, setDescription] = useState('');
  const [format, setFormat] = useState('cklb');
  const [targetData, setTargetData] = useState({
    hostName: '',
    ipAddress: '',
    role: '',
    comments: ''
  });

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();

    onSubmit({
      name: name.trim(),
      description: description.trim(),
      format,
      targetData
    });
  };

  return (
    <div className="dialog-overlay">
      <div className="dialog">
        <h2>Create New Checklist</h2>
        <form onSubmit={handleSubmit}>
          <div className="form-group">
            <label htmlFor="checklist-name">Checklist Name (optional)</label>
            <input
              id="checklist-name"
              type="text"
              value={name}
              onChange={(e) => setName(e.target.value)}
              placeholder="Leave blank to use filename as checklist name"
            />
          </div>

          <div className="form-group">
            <label htmlFor="checklist-description">Description</label>
            <textarea
              id="checklist-description"
              value={description}
              onChange={(e) => setDescription(e.target.value)}
              placeholder="Optional description"
              rows={3}
            />
          </div>


          <div className="form-section">
            <h3>Target System Information</h3>
            <div className="form-row">
              <div className="form-group">
                <label htmlFor="target-hostname">Host Name</label>
                <input
                  id="target-hostname"
                  type="text"
                  value={targetData.hostName}
                  onChange={(e) => setTargetData({ ...targetData, hostName: e.target.value })}
                  placeholder="hostname"
                />
              </div>
              <div className="form-group">
                <label htmlFor="target-ip">IP Address</label>
                <input
                  id="target-ip"
                  type="text"
                  value={targetData.ipAddress}
                  onChange={(e) => setTargetData({ ...targetData, ipAddress: e.target.value })}
                  placeholder="192.168.1.100"
                />
              </div>
            </div>
            <div className="form-group">
              <label htmlFor="target-role">Role</label>
              <input
                id="target-role"
                type="text"
                value={targetData.role}
                onChange={(e) => setTargetData({ ...targetData, role: e.target.value })}
                placeholder="Web Server, Database, etc."
              />
            </div>
            <div className="form-group">
              <label htmlFor="target-comments">Comments</label>
              <textarea
                id="target-comments"
                value={targetData.comments}
                onChange={(e) => setTargetData({ ...targetData, comments: e.target.value })}
                placeholder="Additional system information"
                rows={2}
              />
            </div>
          </div>

          <div className="selected-stigs">
            <h3>Selected STIGs ({selectedStigs.length})</h3>
            <ul>
              {selectedStigs.map((stig) => (
                <li key={stig.id}>
                  <strong>{stig.name}</strong> - {stig.version} ({stig.ruleCount} rules)
                </li>
              ))}
            </ul>
          </div>

          <div className="dialog-actions">
            <button type="button" className="btn btn-secondary" onClick={onCancel}>
              Cancel
            </button>
            <button type="submit" className="btn btn-success">
              Create Checklist
            </button>
          </div>
        </form>
      </div>
    </div>
  );
}

// ChecklistViewer component - Three-panel layout based on STIGViewer3
interface ChecklistViewerProps {
  checklist: any;
  items: any[];
  selectedRule: any;
  onRuleSelect: (rule: any) => void;
  onUpdateRule: (status: string, findingDetails: string, comments: string, ruleId?: string) => void;
  onBack: () => void;
}

function ChecklistViewer({ checklist, items, selectedRule, onRuleSelect, onUpdateRule, onBack }: ChecklistViewerProps) {
  const [statusFilter, setStatusFilter] = useState('all');
  const [severityFilter, setSeverityFilter] = useState('all');
  const [isLoading, setIsLoading] = useState(false);

  // Filter items based on current filters
  const filteredItems = items.filter(item => {
    const statusMatch = statusFilter === 'all' || item.Item.status === statusFilter;
    const severityMatch = severityFilter === 'all' || item.Rule.severity === severityFilter;
    return statusMatch && severityMatch;
  });

  // Calculate summary statistics
  const statusCounts = items.reduce((acc, item) => {
    acc[item.Item.status] = (acc[item.Item.status] || 0) + 1;
    return acc;
  }, {});

  const handleSaveChecklist = async () => {
    if (!checklist) return;

    try {
      setIsLoading(true);

      // Show format selection dialog
      const format = await showFormatDialog();
      if (!format) return; // User cancelled

      const extension = format === 'ckl' ? '.ckl' : '.cklb';
      const defaultName = `${checklist.name}${extension}`;

      const outputPath = await SaveFileDialog(defaultName);
      if (!outputPath) return; // User cancelled

      // Create checklist data structure for saving
      const checklistData = {
        Checklist: checklist,
        Items: items,
        STIGs: items.length > 0 ? [items[0].STIG] : [],
        convertValues: function () { } // Required by the type but not used
      };

      // Save the file
      await SaveChecklistToFile(checklistData, format, outputPath);

      alert(`Checklist saved successfully to: ${outputPath}`);

    } catch (err) {
      alert(`Failed to save checklist: ${err}`);
    } finally {
      setIsLoading(false);
    }
  };

  const showFormatDialog = (): Promise<string | null> => {
    return new Promise((resolve) => {
      const userChoice = prompt(
        'Select output format:\n\n' +
        '1. CKL (XML format)\n' +
        '2. CKLb (JSON format)\n\n' +
        'Enter 1 or 2:'
      );

      if (userChoice === '1') {
        resolve('ckl');
      } else if (userChoice === '2') {
        resolve('cklb');
      } else {
        resolve(null); // User cancelled or invalid input
      }
    });
  };

  return (
    <div className="checklist-viewer-layout">
      <div className="checklist-header">
        <div className="header-actions">
          {/* Relic from a previous iteration
                    <button className="btn btn-secondary" onClick={onBack}>
                        ← Back to Checklists
                    </button> 
                    */}
          <button
            className="btn btn-primary"
            onClick={handleSaveChecklist}
            disabled={isLoading}
          >
            {isLoading ? 'Saving...' : 'Save Checklist'}
          </button>
        </div>
        <h2>{checklist.name}</h2>
        <p>{checklist.description}</p>
      </div>

      <div className="three-panel-layout">
        {/* Left Panel - Rule List */}
        <div className="rule-list-panel">
          <div className="rule-list-header">
            <h3>Rules ({filteredItems.length})</h3>
            <div className="filters">
              <select
                value={statusFilter}
                onChange={(e) => setStatusFilter(e.target.value)}
                className="filter-select"
              >
                <option value="all">All Status</option>
                <option value="NotReviewed">Not Reviewed</option>
                <option value="Open">Open</option>
                <option value="NotAFinding">Not a Finding</option>
                <option value="Not_Applicable">Not Applicable</option>
              </select>
              <select
                value={severityFilter}
                onChange={(e) => setSeverityFilter(e.target.value)}
                className="filter-select"
              >
                <option value="all">All Severity</option>
                <option value="high">High</option>
                <option value="medium">Medium</option>
                <option value="low">Low</option>
              </select>
            </div>
          </div>
          <div className="rule-list">
            {filteredItems.map((item) => (
              <div
                key={item.Item.id}
                className={`rule-item ${selectedRule?.Item.id === item.Item.id ? 'selected' : ''}`}
                onClick={() => onRuleSelect(item)}
              >
                <div className="rule-item-header">
                  <span className={`status-indicator status-${item.Item.status.toLowerCase().replace('_', '-')}`}>
                    {getStatusSymbol(item.Item.status)}
                  </span>
                  <span className="rule-id">{item.Item.vulnId}</span>
                  <span className={`severity severity-${item.Rule.severity?.toLowerCase() || 'unknown'}`}>
                    {item.Rule.severity}
                  </span>
                </div>
                <div className="rule-title">
                  {item.Rule.title || 'No title'}
                </div>
                <div className="rule-stig">
                  {item.STIG.name}
                </div>
              </div>
            ))}
          </div>
        </div>

        {/* Center Panel - Rule Details */}
        <div className="rule-details-panel">
          {selectedRule ? (
            <RuleDetailsPanel
              rule={selectedRule}
              onUpdateRule={onUpdateRule}
            />
          ) : (
            <div className="no-rule-selected">
              <p>Select a rule from the list to view details</p>
            </div>
          )}
        </div>

        {/* Right Panel - Summary Dashboard */}
        <div className="summary-panel">
          <SummaryDashboard
            items={items}
            statusCounts={statusCounts}
          />
        </div>
      </div>
    </div>
  );
}

// Helper function to get status symbols
function getStatusSymbol(status: string): string {
  switch (status) {
    case 'Open': return '●';
    case 'NotAFinding': return '○';
    case 'Not_Applicable': return '◌';
    case 'NotReviewed': return '?';
    default: return '?';
  }
}

// Rule Details Panel Component
interface RuleDetailsPanelProps {
  rule: any;
  onUpdateRule: (status: string, findingDetails: string, comments: string, ruleId?: string) => void;
}

function RuleDetailsPanel({ rule, onUpdateRule }: RuleDetailsPanelProps) {
  const [status, setStatus] = useState(rule.Item.status);
  const [findingDetails, setFindingDetails] = useState(rule.Item.findingDetails ? rule.Item.findingDetails : '');
  const [comments, setComments] = useState(rule.Item.comments ? rule.Item.comments : '');
  const [hasUnsavedChanges, setHasUnsavedChanges] = useState(false);

  // Store previous rule to detect rule changes and current values for auto-save
  const previousRuleRef = React.useRef(rule);
  const currentValuesRef = React.useRef({ status, findingDetails, comments });

  // Update current values ref whenever state changes
  React.useEffect(() => {
    currentValuesRef.current = { status, findingDetails, comments };
  }, [status, findingDetails, comments]);

  // Handle rule changes - auto-save and reset fields
  React.useEffect(() => {
    const previousRule = previousRuleRef.current;

    // Check if we're switching to a different rule
    const previousRuleId = `${previousRule.Item.vulnId}-${previousRule.STIG.id}`;
    const currentRuleId = `${rule.Item.vulnId}-${rule.STIG.id}`;
    if (previousRuleId !== currentRuleId) {
      // Auto-save the previous rule's changes if there were any
      if (hasUnsavedChanges) {
        const { status: prevStatus, findingDetails: prevFindingDetails, comments: prevComments } = currentValuesRef.current;
        const prevRuleId = `${previousRule.Item.vulnId}-${previousRule.STIG.id}`;
        onUpdateRule(prevStatus, prevFindingDetails, prevComments, prevRuleId);
      }

      // Update to new rule's values
      setStatus(rule.Item.status);
      setFindingDetails(rule.Item.findingDetails ? rule.Item.findingDetails : '');
      setComments(rule.Item.comments ? rule.Item.comments : '');
      setHasUnsavedChanges(false);

      // Update the ref for next comparison
      previousRuleRef.current = rule;
    }
  }, [rule.Item.id, hasUnsavedChanges, onUpdateRule]);

  // Check for changes to mark as unsaved
  React.useEffect(() => {
    const hasChanges = (
      status !== rule.Item.status ||
      findingDetails !== (rule.Item.findingDetails ? rule.Item.findingDetails : '') ||
      comments !== (rule.Item.comments ? rule.Item.comments : '')
    );
    setHasUnsavedChanges(hasChanges);
  }, [status, findingDetails, comments, rule.Item.status, rule.Item.findingDetails, rule.Item.comments]);

  const handleSave = () => {
    const ruleId = `${rule.Item.vulnId}-${rule.STIG.id}`;
    onUpdateRule(status, findingDetails, comments, ruleId);
    setHasUnsavedChanges(false);
  };

  return (
    <div className="rule-details">
      <div className="rule-details-header">
        <h3>{rule.Rule.title}</h3>
        <span className="rule-id">{rule.Item.vulnId}</span>
      </div>

      <div className="rule-metadata">
        <div className="metadata-item">
          <label>Severity:</label>
          <span className={`severity severity-${rule.Rule.severity?.toLowerCase() || 'unknown'}`}>
            {rule.Rule.severity}
          </span>
        </div>
        <div className="metadata-item">
          <label>Group:</label>
          <span>{rule.Rule.groupTitle}</span>
        </div>
        <div className="metadata-item">
          <label>STIG:</label>
          <span>{rule.STIG.name}</span>
        </div>
      </div>

      <div className="rule-content">
        <div className="content-section">
          <h4>Discussion</h4>
          <div className="content-text">
            {rule.Rule.description || 'No description available'}
          </div>
        </div>

        <div className="content-section">
          <h4>Check Text</h4>
          <div className="content-text">
            {rule.Rule.checkContent || 'No check content available'}
          </div>
        </div>

        <div className="content-section">
          <h4>Fix Text</h4>
          <div className="content-text">
            {rule.Rule.fixText || 'No fix text available'}
          </div>
        </div>
      </div>

      <div className="finding-details">
        <h4>Finding Details</h4>
        <div className="form-group">
          <label htmlFor="rule-status">Status</label>
          <select
            id="rule-status"
            value={status}
            onChange={(e) => setStatus(e.target.value)}
            className="status-select"
          >
            <option value="NotReviewed">Not Reviewed</option>
            <option value="Open">Open</option>
            <option value="NotAFinding">Not a Finding</option>
            <option value="Not_Applicable">Not Applicable</option>
          </select>
        </div>

        <div className="form-group">
          <label htmlFor="finding-details">Finding Details</label>
          <textarea
            id="finding-details"
            value={findingDetails}
            onChange={(e) => setFindingDetails(e.target.value)}
            placeholder="Enter finding details..."
            rows={4}
          />
        </div>

        <div className="form-group">
          <label htmlFor="comments">Comments</label>
          <textarea
            id="comments"
            value={comments}
            onChange={(e) => setComments(e.target.value)}
            placeholder="Enter comments..."
            rows={3}
          />
        </div>

        <button
          className="btn btn-primary"
          onClick={handleSave}
          disabled={!hasUnsavedChanges}
        >
          {hasUnsavedChanges ? 'Save Changes' : 'No Changes'}
        </button>
      </div>
    </div>
  );
}

// Summary Dashboard Component
interface SummaryDashboardProps {
  items: any[];
  statusCounts: Record<string, number>;
}

function SummaryDashboard({ items, statusCounts }: SummaryDashboardProps) {
  const total = items.length;
  const openCount = statusCounts['Open'] || 0;
  const notAFindingCount = statusCounts['NotAFinding'] || 0;
  const notApplicableCount = statusCounts['Not_Applicable'] || 0;
  const notReviewedCount = statusCounts['NotReviewed'] || 0;

  const completionPercentage = total > 0 ? Math.round(((total - notReviewedCount) / total) * 100) : 0;

  return (
    <div className="summary-dashboard">
      <h3>Summary</h3>

      <div className="completion-ring">
        <div className="ring-container">
          <svg width="120" height="120" className="completion-chart">
            <circle
              cx="60"
              cy="60"
              r="50"
              fill="none"
              stroke="#e5e7eb"
              strokeWidth="10"
            />
            <circle
              cx="60"
              cy="60"
              r="50"
              fill="none"
              stroke="#10b981"
              strokeWidth="10"
              strokeDasharray={`${completionPercentage * 3.14} 314`}
              strokeLinecap="round"
              transform="rotate(-90 60 60)"
            />
          </svg>
          <div className="completion-text">
            <span className="percentage">{completionPercentage}%</span>
            <span className="label">Complete</span>
          </div>
        </div>
      </div>

      <div className="status-breakdown">
        <div className="status-item">
          <span className="status-color status-open"></span>
          <span className="status-label">Open</span>
          <span className="status-count">{openCount}</span>
        </div>
        <div className="status-item">
          <span className="status-color status-not-a-finding"></span>
          <span className="status-label">Not a Finding</span>
          <span className="status-count">{notAFindingCount}</span>
        </div>
        <div className="status-item">
          <span className="status-color status-not-applicable"></span>
          <span className="status-label">Not Applicable</span>
          <span className="status-count">{notApplicableCount}</span>
        </div>
        <div className="status-item">
          <span className="status-color status-not-reviewed"></span>
          <span className="status-label">Not Reviewed</span>
          <span className="status-count">{notReviewedCount}</span>
        </div>
      </div>

      <div className="summary-stats">
        <div className="stat-item">
          <span className="stat-value">{total}</span>
          <span className="stat-label">Total Rules</span>
        </div>
        <div className="stat-item">
          <span className="stat-value">{Math.round((openCount / total) * 100) || 0}%</span>
          <span className="stat-label">Findings</span>
        </div>
      </div>
    </div>
  );
}

// ProgressPopup component
interface ProgressPopupProps {
  progress: ImportProgress;
  onClose: () => void;
}

function ProgressPopup({ progress, onClose }: ProgressPopupProps) {
  const getProgressPercentage = () => {
    if (progress.totalFiles > 0) {
      return Math.round((progress.processedFiles / progress.totalFiles) * 100);
    }
    return 0;
  };

  const getStageText = () => {
    switch (progress.stage) {
      case "extracting": return "Extracting bundle...";
      case "parsing": return "Parsing STIG files...";
      case "storing": return "Storing data...";
      case "complete": return "Import complete!";
      default: return "Processing...";
    }
  };

  return (
    <div className="progress-overlay">
      <div className="progress-popup">
        <div className="progress-header">
          <h3>Importing STIG Bundle</h3>
          {progress.stage !== "complete" && (
            <button className="close-button" onClick={onClose}>×</button>
          )}
        </div>

        <div className="progress-content">
          <div className="progress-stage">
            <div className={`stage-icon ${progress.stage}`}>
              {progress.stage === "complete" ? "✓" :
                progress.stage === "extracting" ? "📦" :
                  progress.stage === "parsing" ? "📄" :
                    progress.stage === "storing" ? "💾" : "⏳"}
            </div>
            <span className="stage-text">{getStageText()}</span>
          </div>

          <div className="progress-message">
            {progress.message}
          </div>

          {progress.currentFile && (
            <div className="current-file">
              <small>Current: {progress.currentFile}</small>
            </div>
          )}

          {progress.totalFiles > 0 && (
            <div className="progress-bar-container">
              <div className="progress-bar">
                <div
                  className="progress-fill"
                  style={{ width: `${getProgressPercentage()}%` }}
                ></div>
              </div>
              <div className="progress-text">
                {progress.processedFiles} / {progress.totalFiles} files
                {progress.processedStigs > 0 && ` (${progress.processedStigs} STIGs)`}
              </div>
            </div>
          )}

          {progress.error && (
            <div className="progress-error">
              <span className="error-icon">⚠️</span>
              <span>{progress.error}</span>
            </div>
          )}
        </div>
      </div>
    </div>
  );
}

export default App;
