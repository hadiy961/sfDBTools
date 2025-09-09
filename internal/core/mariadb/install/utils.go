package install

// utils.go berisi helper functions yang digunakan bersama oleh berbagai modul instalasi
// File ini mengikuti prinsip DRY (Don't Repeat Yourself) dan single responsibility

// File ini sengaja kosong karena helper functions sudah dipindahkan ke file-file yang relevan:
// - isMariaDBInstalled, getInstalledMariaDBVersion, isRunningAsRoot -> precheck.go
// - Helper untuk repo setup -> repo_setup.go
// - Helper untuk package management -> package_install.go
// - Helper untuk service management -> service.go

// Jika ada helper functions yang digunakan di beberapa file,
// letakkan di file ini untuk menghindari duplikasi
