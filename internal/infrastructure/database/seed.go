package database

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

const SeedPassword = "password123"

func seedID(kind, key string) uuid.UUID {
	return uuid.NewSHA1(uuid.NameSpaceOID, []byte(kind+":"+key))
}

func saveAll[T any](tx *gorm.DB, label string, items []T) error {
	if len(items) == 0 {
		return nil
	}
	if err := tx.Save(&items).Error; err != nil {
		return fmt.Errorf("failed to seed %s: %w", label, err)
	}
	return nil
}

func SeedInitialData(db *gorm.DB) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(SeedPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash seed password: %w", err)
	}
	password := string(hash)

	wib := time.FixedZone("WIB", 7*60*60)
	now := time.Now().In(wib)
	at := func(days, hour, minute int) time.Time {
		return time.Date(now.Year(), now.Month(), now.Day()+days, hour, minute, 0, 0, wib)
	}

	const (
		evGit    = "git-workshop"
		evCareer = "career-seminar"
		evHack   = "hackathon"
		evGo     = "golang-workshop"
		evUIUX   = "uiux-bootcamp"
		evAI     = "ai-seminar"
	)

	accounts := []struct{ key, email string }{
		{"rizky", "rizky.pratama@example.com"},
		{"dewi", "dewi.lestari@example.com"},
		{"ahmad", "ahmad.fauzi@example.com"},
		{"siti", "siti.nurhaliza@example.com"},
		{"bagas", "bagas.wicaksono@example.com"},
		{"putri", "putri.ayu@example.com"},
		{"eko", "eko.prasetyo@example.com"},
		{"nadia", "nadia.rahmawati@example.com"},
		{"fajar", "fajar.nugroho@example.com"},
		{"maya", "maya.sari@example.com"},
		{"dimas", "dimas.anggara@example.com"},
		{"laras", "laras.wati@example.com"},
		{"hendra", "hendra.gunawan@example.com"},
		{"indah", "indah.permata@example.com"},
	}

	users := make([]User, 0, len(accounts))
	for i, a := range accounts {
		joined := at(-120+2*i, 8, 0)
		users = append(users, User{
			ID:        seedID("user", a.key),
			Name:      a.key,
			Email:     a.email,
			Password:  password,
			CreatedAt: joined,
			UpdatedAt: joined,
		})
	}

	newEvent := func(key, title, description, location string, date time.Time, maxQuota int, status string, createdDaysAgo int) Event {
		created := at(-createdDaysAgo, 10, 0)
		return Event{
			ID:             seedID("event", key),
			Title:          title,
			Description:    description,
			Date:           date,
			Location:       location,
			AvailableQuota: maxQuota,
			MaxQuota:       maxQuota,
			Status:         status,
			CreatedAt:      created,
			UpdatedAt:      created,
		}
	}

	events := []Event{
		newEvent(evGit, "Workshop Git & GitHub untuk Pemula",
			"Pengenalan version control, branching, dan kolaborasi lewat pull request untuk mahasiswa yang baru mulai ngoding.",
			"Laboratorium Komputer Gedung F, Lantai 2", at(-30, 9, 0), 40, EventStatusFinished, 60),
		newEvent(evCareer, "Seminar Karier di Dunia Teknologi",
			"Praktisi industri berbagi cerita, tips menyusun CV dan portofolio, serta peluang karier bagi lulusan baru di bidang teknologi.",
			"Auditorium Gedung A", at(-14, 13, 0), 150, EventStatusFinished, 45),
		newEvent(evHack, "BCC Hackathon 2026",
			"Kompetisi pengembangan aplikasi selama sehari penuh. Bentuk tim, bangun solusi, dan presentasikan di depan juri.",
			"Gedung Serbaguna Kampus", now.Add(-2*time.Hour), 60, EventStatusOngoing, 30),
		newEvent(evGo, "Workshop Backend Development dengan Golang",
			"Membangun REST API dari nol dengan Go, GORM, dan autentikasi JWT. Peserta wajib membawa laptop sendiri.",
			"Laboratorium Komputer Gedung F, Lantai 3", at(7, 9, 0), 30, EventStatusUpcoming, 12),
		newEvent(evUIUX, "Bootcamp UI/UX Design",
			"Belajar prinsip desain antarmuka, wireframing, prototyping di Figma, hingga usability testing bersama mentor berpengalaman.",
			"Ruang Seminar Gedung B, Lantai 4", at(14, 8, 30), 8, EventStatusUpcoming, 14),
		newEvent(evAI, "Seminar Nasional Kecerdasan Buatan",
			"Seminar tentang perkembangan AI di Indonesia, penerapan large language model, serta etika dan regulasinya.",
			"Auditorium Gedung A", at(30, 8, 0), 200, EventStatusUpcoming, 10),
	}

	type registrationSeed struct{ event, user, status string }
	var registrations []registrationSeed
	register := func(event, status string, userKeys ...string) {
		for _, u := range userKeys {
			registrations = append(registrations, registrationSeed{event, u, status})
		}
	}

	register(evGit, RegistrationStatusCheckedIn, "putri", "eko", "nadia", "fajar", "maya", "dimas")
	register(evGit, RegistrationStatusRegistered, "laras")
	register(evGit, RegistrationStatusCancelled, "hendra")

	register(evCareer, RegistrationStatusCheckedIn, "putri", "eko", "nadia", "fajar", "maya", "laras", "indah")
	register(evCareer, RegistrationStatusRegistered, "dimas")

	register(evHack, RegistrationStatusCheckedIn, "putri", "eko", "nadia", "fajar", "maya")
	register(evHack, RegistrationStatusRegistered, "dimas", "laras", "hendra", "indah")

	register(evGo, RegistrationStatusRegistered, "putri", "nadia", "fajar", "maya", "dimas", "hendra", "indah")

	register(evUIUX, RegistrationStatusRegistered, "eko", "nadia", "fajar", "maya", "dimas", "laras", "hendra", "indah")
	register(evUIUX, RegistrationStatusCancelled, "putri")

	register(evAI, RegistrationStatusRegistered, "putri", "eko", "nadia", "fajar", "maya", "laras", "indah")

	active := make(map[uuid.UUID]int)
	for _, r := range registrations {
		if r.status != RegistrationStatusCancelled {
			active[seedID("event", r.event)]++
		}
	}

	eventsByID := make(map[uuid.UUID]Event, len(events))
	for i := range events {
		events[i].AvailableQuota = events[i].MaxQuota - active[events[i].ID]
		eventsByID[events[i].ID] = events[i]
	}

	registrationRows := make([]Registration, 0, len(registrations))
	for i, r := range registrations {
		event := eventsByID[seedID("event", r.event)]
		created := event.CreatedAt.Add(48*time.Hour + time.Duration(i)*time.Hour)
		updated := created
		switch r.status {
		case RegistrationStatusCheckedIn:
			updated = event.Date.Add(15 * time.Minute)
		case RegistrationStatusCancelled:
			updated = created.Add(24 * time.Hour)
		}
		registrationRows = append(registrationRows, Registration{
			ID:        seedID("registration", r.event+":"+r.user),
			EventID:   event.ID,
			UserID:    seedID("user", r.user),
			Status:    r.status,
			CreatedAt: created,
			UpdatedAt: updated,
		})
	}

	committeePlan := []struct {
		event       string
		coordinator string
		staff       []string
	}{
		{evGit, "rizky", []string{"dewi"}},
		{evCareer, "dewi", []string{"ahmad", "siti"}},
		{evHack, "ahmad", []string{"rizky", "bagas", "siti"}},
		{evGo, "rizky", []string{"bagas"}},
		{evUIUX, "siti", []string{"dewi"}},
		{evAI, "bagas", []string{"ahmad", "rizky"}},
	}

	var committees []Committee
	var invitations []Invitation
	for _, plan := range committeePlan {
		eventID := seedID("event", plan.event)
		created := eventsByID[eventID].CreatedAt

		committees = append(committees, Committee{
			ID:        seedID("committee", plan.event+":"+plan.coordinator),
			EventID:   eventID,
			UserID:    seedID("user", plan.coordinator),
			Role:      RoleCoordinator,
			CreatedAt: created,
			UpdatedAt: created,
		})

		for i, staff := range plan.staff {
			invited := created.Add(time.Duration(i+1) * 2 * time.Hour)
			accepted := invited.Add(3 * time.Hour)

			invitations = append(invitations, Invitation{
				ID:         seedID("invitation", plan.event+":"+staff),
				EventID:    eventID,
				SenderID:   seedID("user", plan.coordinator),
				ReceiverID: seedID("user", staff),
				Role:       RoleStaff,
				Status:     InvitationStatusAccepted,
				CreatedAt:  invited,
				UpdatedAt:  accepted,
			})
			committees = append(committees, Committee{
				ID:        seedID("committee", plan.event+":"+staff),
				EventID:   eventID,
				UserID:    seedID("user", staff),
				Role:      RoleStaff,
				CreatedAt: accepted,
				UpdatedAt: accepted,
			})
		}
	}

	extraInvitations := []struct {
		event, sender, receiver, role, status string
		daysAgo                               int
	}{
		{evGo, "rizky", "eko", RoleStaff, InvitationStatusPending, 2},
		{evGo, "rizky", "dewi", RoleCoordinator, InvitationStatusPending, 1},
		{evUIUX, "siti", "putri", RoleStaff, InvitationStatusDeclined, 8},
		{evAI, "bagas", "hendra", RoleStaff, InvitationStatusCancelled, 5},
	}
	for _, inv := range extraInvitations {
		created := at(-inv.daysAgo, 11, 0)
		updated := created
		if inv.status != InvitationStatusPending {
			updated = created.Add(6 * time.Hour)
		}
		invitations = append(invitations, Invitation{
			ID:         seedID("invitation", inv.event+":"+inv.receiver),
			EventID:    seedID("event", inv.event),
			SenderID:   seedID("user", inv.sender),
			ReceiverID: seedID("user", inv.receiver),
			Role:       inv.role,
			Status:     inv.status,
			CreatedAt:  created,
			UpdatedAt:  updated,
		})
	}

	dateOf := func(event string) time.Time {
		return eventsByID[seedID("event", event)].Date
	}
	newSession := func(event, key, title, speaker, location string, startsAt time.Time, status string) Session {
		eventID := seedID("event", event)
		created := eventsByID[eventID].CreatedAt.Add(24 * time.Hour)
		return Session{
			ID:        seedID("session", event+":"+key),
			EventID:   eventID,
			Title:     title,
			Time:      startsAt,
			Speaker:   speaker,
			Location:  location,
			Status:    status,
			CreatedAt: created,
			UpdatedAt: created,
		}
	}

	const (
		labF2 = "Laboratorium Komputer Gedung F, Lantai 2"
		labF3 = "Laboratorium Komputer Gedung F, Lantai 3"
		audA  = "Auditorium Gedung A"
		gsk   = "Gedung Serbaguna Kampus"
		roomB = "Ruang Seminar Gedung B, Lantai 4"
	)

	sessions := []Session{
		newSession(evGit, "intro", "Pengenalan Version Control dan Git", "Rizky Pratama", labF2, dateOf(evGit), SessionStatusFinished),
		newSession(evGit, "branching", "Praktik Branching, Merge, dan Resolusi Konflik", "Dewi Lestari", labF2, dateOf(evGit).Add(90*time.Minute), SessionStatusFinished),
		newSession(evGit, "pull-request", "Kolaborasi Tim dengan GitHub Pull Request", "Rizky Pratama", labF2, dateOf(evGit).Add(4*time.Hour), SessionStatusFinished),

		newSession(evCareer, "landscape", "Lanskap Karier Software Engineer di Indonesia", "Andi Kurniawan", audA, dateOf(evCareer), SessionStatusFinished),
		newSession(evCareer, "panel", "Panel Diskusi: Dari Kampus ke Industri", "Andi Kurniawan, Maria Susanti, dan Yoga Permana", audA, dateOf(evCareer).Add(105*time.Minute), SessionStatusFinished),
		newSession(evCareer, "portfolio", "Menyiapkan CV dan Portofolio yang Menarik", "Maria Susanti", "Ruang Diskusi Gedung A", dateOf(evCareer).Add(210*time.Minute), SessionStatusFinished),

		newSession(evHack, "opening", "Pembukaan dan Technical Briefing", "Ahmad Fauzi", gsk, dateOf(evHack), SessionStatusFinished),
		newSession(evHack, "hacking", "Hacking Time", "Panitia BCC Hackathon", gsk, dateOf(evHack).Add(30*time.Minute), SessionStatusOngoing),
		newSession(evHack, "mentoring", "Sesi Mentoring", "Yoga Permana dan Rina Handayani", "Ruang Mentoring Lantai 2", dateOf(evHack).Add(5*time.Hour), SessionStatusUpcoming),
		newSession(evHack, "pitching", "Final Pitching dan Pengumuman Pemenang", "Dewan Juri", gsk, dateOf(evHack).Add(10*time.Hour), SessionStatusUpcoming),

		newSession(evGo, "intro", "Pengenalan Go dan Ekosistemnya", "Rizky Pratama", labF3, dateOf(evGo), SessionStatusUpcoming),
		newSession(evGo, "rest-api", "Membangun REST API dengan Go", "Bagas Wicaksono", labF3, dateOf(evGo).Add(90*time.Minute), SessionStatusUpcoming),
		newSession(evGo, "gorm", "Akses Database dengan GORM", "Rizky Pratama", labF3, dateOf(evGo).Add(4*time.Hour), SessionStatusUpcoming),
		newSession(evGo, "jwt", "Autentikasi JWT dan Deployment", "Bagas Wicaksono", labF3, dateOf(evGo).Add(6*time.Hour), SessionStatusUpcoming),

		newSession(evUIUX, "principles", "Prinsip Dasar UI/UX Design", "Rina Handayani", roomB, dateOf(evUIUX), SessionStatusUpcoming),
		newSession(evUIUX, "figma", "Wireframing dan Prototyping dengan Figma", "Rina Handayani", roomB, dateOf(evUIUX).Add(2*time.Hour), SessionStatusUpcoming),
		newSession(evUIUX, "usability", "Usability Testing", "Siti Nurhaliza", roomB, dateOf(evUIUX).Add(5*time.Hour), SessionStatusUpcoming),

		newSession(evAI, "keynote", "Keynote: Masa Depan Kecerdasan Buatan di Indonesia", "Prof. Hari Setiawan", audA, dateOf(evAI), SessionStatusUpcoming),
		newSession(evAI, "llm", "Large Language Model dalam Praktik", "Maria Susanti", audA, dateOf(evAI).Add(2*time.Hour), SessionStatusUpcoming),
		newSession(evAI, "ethics", "Etika dan Regulasi Kecerdasan Buatan", "Dr. Laila Safitri", audA, dateOf(evAI).Add(4*time.Hour), SessionStatusUpcoming),
		newSession(evAI, "qna", "Sesi Tanya Jawab dan Networking", "Seluruh Pembicara", audA, dateOf(evAI).Add(6*time.Hour), SessionStatusUpcoming),
	}

	announcementPlan := []struct {
		event, sender, title, body string
		when                       time.Time
	}{
		{evGit, "rizky", "Materi Workshop Sudah Tersedia",
			"Slide dan repositori latihan dapat diakses melalui tautan yang dikirim ke email peserta. Terima kasih sudah hadir!", at(-29, 10, 0)},
		{evCareer, "dewi", "Terima Kasih Telah Hadir",
			"Terima kasih kepada seluruh peserta dan pembicara. Rekaman sesi panel akan dibagikan dalam beberapa hari ke depan.", at(-13, 9, 0)},
		{evHack, "ahmad", "Technical Briefing dan Aturan Lomba",
			"Pastikan seluruh anggota tim sudah terdaftar. Repositori final harus dikumpulkan sebelum sesi Final Pitching dimulai.", at(-3, 19, 0)},
		{evHack, "siti", "Jadwal Mentoring Hari Ini",
			"Sesi mentoring dibuka setelah makan siang di Ruang Mentoring Lantai 2. Silakan antre sesuai urutan tim.", now.Add(-90 * time.Minute)},
		{evGo, "rizky", "Persiapan Instalasi Go",
			"Mohon instal Go versi terbaru dan editor pilihan sebelum hari acara agar sesi praktik berjalan lancar.", at(-3, 10, 0)},
		{evGo, "bagas", "Jangan Lupa Membawa Laptop",
			"Peserta diwajibkan membawa laptop pribadi dengan baterai terisi penuh karena colokan listrik terbatas.", at(-1, 15, 30)},
		{evUIUX, "siti", "Kuota Telah Penuh",
			"Seluruh kursi Bootcamp UI/UX sudah terisi. Terima kasih atas antusiasme kalian!", at(-4, 11, 0)},
		{evAI, "bagas", "Pendaftaran Telah Dibuka",
			"Pendaftaran Seminar Nasional Kecerdasan Buatan sudah dibuka. Amankan kursimu sebelum kuota habis.", at(-9, 9, 0)},
		{evAI, "ahmad", "Pembicara Utama Diumumkan",
			"Keynote akan dibawakan oleh Prof. Hari Setiawan. Pantau terus pengumuman selanjutnya.", at(-2, 14, 0)},
	}

	announcements := make([]Announcement, 0, len(announcementPlan))
	for _, a := range announcementPlan {
		announcements = append(announcements, Announcement{
			ID:        seedID("announcement", a.event+":"+a.title),
			Title:     a.title,
			Body:      a.body,
			EventID:   seedID("event", a.event),
			SenderID:  seedID("user", a.sender),
			CreatedAt: a.when,
			UpdatedAt: a.when,
		})
	}

	return db.Transaction(func(tx *gorm.DB) error {
		if err := saveAll(tx, "users", users); err != nil {
			return err
		}
		if err := saveAll(tx, "events", events); err != nil {
			return err
		}
		if err := saveAll(tx, "committees", committees); err != nil {
			return err
		}
		if err := saveAll(tx, "invitations", invitations); err != nil {
			return err
		}
		if err := saveAll(tx, "registrations", registrationRows); err != nil {
			return err
		}
		if err := saveAll(tx, "sessions", sessions); err != nil {
			return err
		}
		return saveAll(tx, "announcements", announcements)
	})
}
