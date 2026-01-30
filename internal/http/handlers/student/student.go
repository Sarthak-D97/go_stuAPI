package student

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/Sarthak-D97/go_stuAPI/internal/storage"
	"github.com/Sarthak-D97/go_stuAPI/internal/types"
	"github.com/Sarthak-D97/go_stuAPI/internal/utils/response"
	"github.com/go-playground/validator/v10"
	"github.com/redis/go-redis/v9"
)

const (
	studentKeyPrefix = "student:"
	studentListKey   = "students_list"
	cacheTTL         = 10 * time.Minute
)

func New(storage storage.Storage, rdb *redis.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		slog.Info("creating a student")
		var student types.Student
		err := json.NewDecoder(r.Body).Decode(&student)
		if errors.Is(err, io.EOF) {
			response.WriteJson(w, http.StatusBadRequest, response.GeneralError(fmt.Errorf("Empty Body")))
			return
		}
		if err != nil {
			response.WriteJson(w, http.StatusBadRequest, response.GeneralError(err))
			return
		}
		if err := validator.New().Struct(student); err != nil {
			validateErr := err.(validator.ValidationErrors)
			response.WriteJson(w, http.StatusBadRequest, response.ValidationError(validateErr))
			return
		}

		lastid, err := storage.CreateStudent(student.Name, student.Email, student.Age)
		if err != nil {
			response.WriteJson(w, http.StatusInternalServerError, response.GeneralError(err))
			return
		}
		student.ID = int(lastid)
		go func() {
			ctx := context.Background()
			cacheKey := fmt.Sprintf("%s%d", studentKeyPrefix, lastid)
			pipe := rdb.Pipeline()
			pipe.HSet(ctx, cacheKey, &student)
			pipe.Expire(ctx, cacheKey, cacheTTL)
			pipe.Del(ctx, studentListKey)
			_, err := pipe.Exec(ctx)

			if err != nil {
				slog.Error("failed to cache student", "error", err)
			}
		}()

		slog.Info("student created successfully", slog.Int64("student_id", lastid))
		response.WriteJson(w, http.StatusCreated, map[string]interface{}{
			"status":     response.StatusCreated,
			"student_id": lastid,
		})
	}
}

func GetById(storage storage.Storage, rdb *redis.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		slog.Info("getting a student by id", slog.String("id", id))
		intId, err := strconv.ParseInt(id, 10, 64)
		if err != nil {
			response.WriteJson(w, http.StatusBadRequest, response.GeneralError(err))
			return
		}
		cacheKey := fmt.Sprintf("%s%s", studentKeyPrefix, id)
		var cachedStudent types.Student
		err = rdb.HGetAll(r.Context(), cacheKey).Scan(&cachedStudent)
		if err == nil && cachedStudent.ID != 0 {
			slog.Info("serving student from cache (hash)", slog.String("id", id))
			response.WriteJson(w, http.StatusOK, map[string]interface{}{
				"status":  response.StatusOK,
				"student": cachedStudent,
			})
			return
		}
		student, err := storage.GetStudentById(intId)
		if err != nil {
			response.WriteJson(w, http.StatusInternalServerError, response.GeneralError(err))
			return
		}
		go func() {
			ctx := context.Background()
			pipe := rdb.Pipeline()
			pipe.HSet(ctx, cacheKey, student)
			pipe.Expire(ctx, cacheKey, cacheTTL)
			if _, err := pipe.Exec(ctx); err != nil {
				slog.Error("failed to cache student", "error", err)
			}
		}()

		slog.Info("student fetched successfully", slog.String("student_id", id))
		response.WriteJson(w, http.StatusOK, map[string]interface{}{
			"status":  response.StatusOK,
			"student": student,
		})
	}
}

func GetList(storage storage.Storage, rdb *redis.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		slog.Info("getting list of students")

		val, err := rdb.Get(r.Context(), studentListKey).Result()
		if err == nil {
			var cachedStudents []types.Student
			if jsonErr := json.Unmarshal([]byte(val), &cachedStudents); jsonErr == nil {
				slog.Info("serving student list from cache")
				response.WriteJson(w, http.StatusOK, map[string]interface{}{
					"status":   response.StatusOK,
					"students": cachedStudents,
				})
				return
			}
		}

		students, err := storage.GetAllStudents()
		if err != nil {
			response.WriteJson(w, http.StatusInternalServerError, err)
			return
		}

		go func() {
			data, _ := json.Marshal(students)
			rdb.Set(context.Background(), studentListKey, data, cacheTTL)
		}()

		slog.Info("students fetched successfully", slog.Int("count", len(students)))
		response.WriteJson(w, http.StatusOK, map[string]interface{}{
			"status":   response.StatusOK,
			"students": students,
		})
	}
}

func UpdateStudent(storage storage.Storage, rdb *redis.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		slog.Info("Updating student", slog.String("id", id))

		intId, err := strconv.ParseInt(id, 10, 64)
		if err != nil {
			response.WriteJson(w, http.StatusBadRequest, response.GeneralError(err))
			return
		}

		student, err := storage.GetStudentById(intId)
		if err != nil {
			response.WriteJson(w, http.StatusInternalServerError, response.GeneralError(err))
			return
		}

		err = json.NewDecoder(r.Body).Decode(&student)
		if err != nil && !errors.Is(err, io.EOF) {
			response.WriteJson(w, http.StatusBadRequest, response.GeneralError(err))
			return
		}

		if err := validator.New().Struct(student); err != nil {
			validateErr := err.(validator.ValidationErrors)
			response.WriteJson(w, http.StatusBadRequest, response.ValidationError(validateErr))
			return
		}

		err = storage.UpdateStudent(intId, student.Name, student.Email, student.Age)
		if err != nil {
			response.WriteJson(w, http.StatusInternalServerError, response.GeneralError(err))
			return
		}

		go func() {
			ctx := context.Background()
			cacheKey := fmt.Sprintf("%s%d", studentKeyPrefix, intId)

			pipe := rdb.Pipeline()
			pipe.HSet(ctx, cacheKey, student)
			pipe.Expire(ctx, cacheKey, cacheTTL)
			pipe.Del(ctx, studentListKey)
			if _, err := pipe.Exec(ctx); err != nil {
				slog.Error("failed to update student cache", "error", err)
			}
		}()

		slog.Info("student updated successfully", slog.Int64("student_id", intId))
		response.WriteJson(w, http.StatusOK, map[string]interface{}{
			"status":     response.StatusOK,
			"student_id": intId,
		})
	}
}
func DeleteStudent(storage storage.Storage, rdb *redis.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		slog.Info("Deleting student", slog.String("id", id))

		intId, err := strconv.ParseInt(id, 10, 64)
		if err != nil {
			response.WriteJson(w, http.StatusBadRequest, response.GeneralError(err))
			return
		}

		err = storage.DeleteStudent(intId)
		if err != nil {
			response.WriteJson(w, http.StatusInternalServerError, response.GeneralError(err))
			return
		}

		go func() {
			ctx := context.Background()
			pipe := rdb.Pipeline()
			pipe.Del(ctx, fmt.Sprintf("%s%s", studentKeyPrefix, id))
			pipe.Del(ctx, studentListKey)
			pipe.Exec(ctx)
		}()

		slog.Info("student deleted successfully", slog.String("student_id", id))
		response.WriteJson(w, http.StatusOK, map[string]interface{}{
			"status": response.StatusOK,
			"msg":    "student deleted successfully",
		})
	}
}
