# Encryption Consistency in sfDBTools

## Overview
sfDBTools now uses a unified encryption approach across all components to ensure consistency between configuration files and backup files.

## Encryption Method
- **Algorithm**: AES-256-GCM (authenticated encryption)
- **Key Derivation**: PBKDF2 with SHA-512
- **Salt**: Fixed salt "sfdb_encryption_salt_v3" for consistency
- **Iterations**: 10,000 iterations (configurable)

## Components Using Unified Encryption

### 1. Config Generation (`config generate`)
- Uses user-provided encryption password
- Derives key using `crypto.DeriveKeyWithPassword(userPassword)`
- Saves encrypted database configuration (.cnf.enc files)

### 2. Backup Encryption (`backup all`, `backup selection`, etc.)
- Uses the same user-provided encryption password
- Derives key using `crypto.DeriveKeyWithPassword(userPassword)`
- Encrypts backup files with .enc extension

### 3. Restore Decryption (`restore all`, `restore single`)
- Prompts for the same encryption password
- Derives key using `crypto.DeriveKeyWithPassword(userPassword)`
- Decrypts backup files encrypted with the backup process

## Password Handling
- **Interactive Mode**: Users are prompted to enter encryption password
- **Environment Variable**: Set `SFDB_ENCRYPTION_PASSWORD` to avoid prompts
- **Consistency**: Same password works for both config files and backup files

## Migration from Old Encryption
If you have backup files encrypted with the old method (using app config values), you will need to:
1. Decrypt old backup files using the old decryption method
2. Re-encrypt them using the new unified encryption method
3. Or use the migration commands when available

## Security Benefits
1. **User Control**: Encryption password is controlled by the user
2. **Consistency**: One password for all encrypted data
3. **Portability**: Encrypted files can be moved between environments
4. **Authentication**: GCM mode provides built-in authentication

## Usage Examples

### Config Generation
```bash
# Generate encrypted config
sfDBTools config generate --auto --config-name mydb --db-host localhost --db-port 3306 --db-user root
# Will prompt for encryption password

# Or use environment variable
export SFDB_ENCRYPTION_PASSWORD="your_secret_password"
sfDBTools config generate --auto --config-name mydb --db-host localhost --db-port 3306 --db-user root
```

### Backup with Encryption
```bash
# Backup with encryption (will prompt for password)
sfDBTools backup all --source_host localhost --source_user root --encrypt

# Or use environment variable
export SFDB_ENCRYPTION_PASSWORD="your_secret_password"
sfDBTools backup all --source_host localhost --source_user root --encrypt
```

### Restore Encrypted Backup
```bash
# Restore encrypted backup (will prompt for password)
sfDBTools restore all --target_host localhost --target_user root --file backup_file.sql.gz.enc

# Or use environment variable
export SFDB_ENCRYPTION_PASSWORD="your_secret_password"
sfDBTools restore all --target_host localhost --target_user root --file backup_file.sql.gz.enc
```

## Error Messages
- If wrong password is provided: "failed to decrypt data (incorrect password or data corruption)"
- If password is empty: "password cannot be empty"
- If file is corrupted: Authentication will fail during GCM decryption

## Implementation Details
- Key derivation function: `crypto.DeriveKeyWithPassword()`
- Salt: `"sfdb_encryption_salt_v3"`
- Key length: 32 bytes (AES-256)
- File format: Nonce (12 bytes) + Ciphertext + AuthTag (16 bytes)
